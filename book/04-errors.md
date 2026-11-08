# 第4章 エラーハンドリング

Go には例外 (`try/catch`) がありません。エラーは **戻り値として返す普通の値** です (*Errors are values*)。

## 4.1 基本形

```go
f, err := os.Open(path)
if err != nil {
	return fmt.Errorf("open config: %w", err)
}
defer f.Close()
```

- エラーは **最後の戻り値**、型は組み込みインターフェース `error`。
- 呼び出し直後に `if err != nil` でチェックし、**早期リターン** する。正常系のコードはインデントが浅いまま左端を流れる ("happy path を左に寄せる")。

```go
type error interface {
	Error() string
}
```

## 4.2 ラップ (`%w`) と判定 (`errors.Is` / `errors.As`)

Go 1.13 以降、エラーに文脈を足しつつ元のエラーを保持できます。

```go
var ErrNotFound = errors.New("not found") // センチネルエラー

func findUser(id int) (*User, error) {
	// ...
	return nil, fmt.Errorf("find user id=%d: %w", id, ErrNotFound)
}

_, err := findUser(42)
fmt.Println(err)                         // find user id=42: not found
fmt.Println(errors.Is(err, ErrNotFound)) // true (ラップを辿って比較)
```

独自のエラー型で詳細情報を持たせる場合:

```go
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

var verr *ValidationError
if errors.As(err, &verr) { // ラップを辿って型が一致するものを取り出す
	fmt.Println("invalid field:", verr.Field)
}
```

| 方法 | 使いどころ |
| --- | --- |
| `errors.New` のセンチネル (`ErrXxx`) | 呼び出し側が「種類」だけ判定したい (`io.EOF`, `sql.ErrNoRows`, `fs.ErrNotExist`) |
| 独自エラー型 (`XxxError`) | 追加情報 (フィールド名、ステータスコード等) を渡したい |
| `fmt.Errorf("...: %w", err)` | 文脈を付けて上に返す |
| `fmt.Errorf("...: %v", err)` | 内部エラーを API として公開したくない (判定させない) |
| `errors.Join(e1, e2)` (1.20+) | 複数のエラーをまとめる |

### メッセージの書き方

[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#error-strings) の慣習:

- **小文字で始め、句点を付けない** (`"open config: permission denied"` のように連結されるため)
- `"failed to ..."` を毎回付けない。何をしていたか (`"open config"`) だけで十分

## 4.3 panic と recover

`panic` は **プログラムのバグ** (配列外参照、nil 参照、到達しないはずの分岐) のためのものです。通常のエラー処理には使いません。

```go
func mustParse(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err) // Must 系関数: 初期化時など失敗=バグの場面でだけ使う
	}
	return n
}
```

HTTP サーバーでは、1 リクエストの panic でプロセス全体が落ちないよう `recover` するミドルウェアを入れるのが定番です (第11章で実装)。

## 4.4 エラーはログに出すか、返すか — どちらか一方

```go
// ❌ ログを出したうえで返す → 上位でもログが出て二重になる
if err != nil {
	log.Printf("query failed: %v", err)
	return err
}

// ✅ 文脈を付けて返し、最上位 (main やハンドラ) で一度だけ処理する
if err != nil {
	return fmt.Errorf("query users: %w", err)
}
```

> 📦 **OSS ではこう書かれている**
>
> - **Kubernetes**: `k8s.io/apimachinery/pkg/api/errors` に `IsNotFound(err)` / `IsConflict(err)` のような判定関数を用意し、呼び出し側はエラー型を直接触らない。
> - **GitHub CLI (cli/cli)**: `cmdutil.SilentError` や `cmdutil.FlagErrorf` のような「表示方法を決めるエラー型」を定義し、最上位の `ghcmd.Main()` でまとめて終了コードとメッセージを決めている。
> - **golangci-lint の `errorlint`**: `err == ErrX` や `err.(*T)` のようにラップを考慮しない比較を検出してくれる。本リポジトリの `.golangci.yml` でも有効化している。

## ✍️ 練習問題

1. ファイルを読んで行数を返す `CountLines(path string) (int, error)` を書き、ファイルが存在しない場合に呼び出し側で `errors.Is(err, fs.ErrNotExist)` で判定してみよう。
2. `HTTPError{StatusCode int}` 型を作り、`errors.As` で取り出してステータスコードを表示しよう。
