# 第7章 エラーハンドリング

Go には例外 (`try` / `catch`) がありません。エラーは **戻り値として返す普通の値** です。

> *Errors are values.* — Rob Pike

## 7.1 error インターフェース

```go
type error interface {
	Error() string
}
```

エラーを返す可能性のある関数は、**最後の戻り値** を `error` 型にします。呼び出し側はすぐに `nil` かどうかを確認します。

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	i, err := strconv.Atoi("42a")
	if err != nil {
		fmt.Println("couldn't convert number:", err)
		return
	}
	fmt.Println("Converted integer:", i)
}
```

```text
couldn't convert number: strconv.Atoi: parsing "42a": invalid syntax
```

`if err != nil { return ... }` が何度も続くのは冗長に見えるかもしれません。しかしこの書き方には、**どこで失敗しうるかがコードに明示され、エラーを処理し忘れると目立つ** という利点があります。

## 7.2 エラーを作る

```go
package main

import (
	"errors"
	"fmt"
	"time"
)

// 1. errors.New: 固定のメッセージだけを持つエラー
var ErrEmpty = errors.New("empty input")

// 2. fmt.Errorf: 値を埋め込んだメッセージを作る
func parseAge(n int) error {
	if n < 0 {
		return fmt.Errorf("invalid age %d: must be >= 0", n)
	}
	return nil
}

// 3. 独自のエラー型: 追加の情報をフィールドに持たせる
type MyError struct {
	When time.Time
	What string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("at %s, %s", e.When.Format("2006-01-02"), e.What)
}

func run() error {
	return &MyError{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "it didn't work"}
}

func main() {
	fmt.Println(ErrEmpty)
	fmt.Println(parseAge(-1))
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
```

```text
empty input
invalid age -1: must be >= 0
at 2026-01-01, it didn't work
```

標準ライブラリの `os.PathError` は、独自のエラー型の代表例です。どの操作が・どのファイルで・なぜ失敗したかを、フィールドとして持っています。

```go
// PathError records an error and the operation and file path that caused it.
type PathError struct {
	Op   string // "open", "unlink", etc.
	Path string // The associated file.
	Err  error  // Returned by the system call.
}

func (e *PathError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }
// 例: "open /etc/passwx: no such file or directory"
```

### エラーメッセージの書き方

[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#error-strings) の慣習は次のとおりです。

- **小文字で始め、末尾に句点やピリオドを付けません**。エラーメッセージは `"read config: open app.yaml: permission denied"` のように連結されていくからです。
- 何をしようとして失敗したのかが分かるように、**操作の名前や対象を前に付けます** (`"open config: ..."`)。
- `"failed to ..."` や `"error: ..."` を毎回付ける必要はありません。エラーであることは文脈から明らかです。

## 7.3 ラップと判定 — `%w`、errors.Is、errors.As

エラーを上位の関数へ返すときは、**文脈を付け加えつつ、元のエラーも保持** しておくと便利です (Go 1.13+)。

```go
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

var ErrNotFound = errors.New("not found") // センチネルエラー (判定に使う決まった値)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string { return "invalid " + e.Field }

func findUser(id int) error {
	if id == 0 {
		return &ValidationError{Field: "id"}
	}
	return fmt.Errorf("find user id=%d: %w", id, ErrNotFound) // %w でラップする
}

func handler(id int) error {
	if err := findUser(id); err != nil {
		return fmt.Errorf("handler: %w", err) // ラップを何段重ねてもよい
	}
	return nil
}

func main() {
	err := handler(42)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrNotFound)) // ラップをたどって一致するか確かめる

	err = handler(0)
	var verr *ValidationError
	if errors.As(err, &verr) { // ラップをたどって、型が一致するエラーを取り出す
		fmt.Println("validation failed on:", verr.Field)
	}

	_, err = os.Open("/no/such/file")
	fmt.Println(errors.Is(err, fs.ErrNotExist))
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		fmt.Println("op:", pathErr.Op, "path:", pathErr.Path)
	}

	// 複数のエラーをまとめる (Go 1.20+)
	joined := errors.Join(ErrNotFound, &ValidationError{"name"})
	fmt.Println(errors.Is(joined, ErrNotFound))
	fmt.Println(joined)
}
```

```text
handler: find user id=42: not found
true
validation failed on: id
true
op: open path: /no/such/file
true
not found
invalid name
```

| やりたいこと | 方法 |
| --- | --- |
| 呼び出し側でエラーの **種類** を判定させたい | センチネルエラー `var ErrXxx = errors.New(...)` を用意し、`errors.Is` で判定する (`io.EOF`、`sql.ErrNoRows`、`fs.ErrNotExist`) |
| エラーに **追加の情報** を持たせたい | 独自のエラー型 `XxxError` を作り、`errors.As` で取り出す |
| 文脈を付けて上に返し、元のエラーも判定できるようにしたい | `fmt.Errorf("...: %w", err)` |
| 文脈は付けたいが、内部のエラーは **公開 API の一部にしたくない** | `fmt.Errorf("...: %v", err)` (ラップしないので判定できない) |
| 複数のエラーをまとめたい | `errors.Join` (`%w` を 2 回以上使った `fmt.Errorf` でもよい) |

> ⚠️ `err == ErrNotFound` や `err.(*ValidationError)` のように直接比較すると、**ラップされたエラーは見落とされます**。必ず `errors.Is` と `errors.As` を使いましょう (golangci-lint の `errorlint` で検出できます)。

> 💡 ラップした時点で、そのエラーは **パッケージの API の一部** になります。利用者が `errors.Is(err, sql.ErrNoRows)` に頼り始めると、後で別のデータベースに切り替えたときに互換性が壊れます。外に見せたくない実装の詳細は、`%v` で文字列として埋め込むだけにしましょう。

## 7.4 エラーは 1 回だけ処理する

```go
// ❌ ログを出したうえでエラーも返す → 上位でもログが出て、同じエラーが何度も記録される
if err != nil {
	log.Printf("query failed: %v", err)
	return err
}

// ✅ 文脈を付けて返すだけにして、最上位 (main や HTTP ハンドラ) で 1 回だけログを出す
if err != nil {
	return fmt.Errorf("query users: %w", err)
}
```

エラーへの対応は、**処理する (ログに出す、リトライする、代わりの値を使う)** か **返す** かのどちらか一方にします。

## 7.5 panic

`panic` は、通常の制御の流れを止める組み込み関数です。panic が起きると、その関数の `defer` が実行されてから呼び出し元へ戻り、呼び出し元でも同じように `defer` が実行されます。これを繰り返して最上位まで達すると、プログラムはスタックトレースを表示して終了します。

panic は **プログラマのミス (バグ) や、続行が不可能な状態** のためのものです。

- 範囲外の添字アクセス、nil ポインタの参照、nil マップへの書き込み → ランタイムが panic を起こします。
- 「ここには絶対に来ないはず」の分岐。
- **初期化時の失敗**。プログラムの起動に必要な設定が読めないなど。

```go
var user = os.Getenv("USER")

func init() {
	if user == "" {
		panic("no value for $USER")
	}
}
```

`regexp.MustCompile` や `template.Must` のように **`Must` で始まる関数** は、失敗すると panic します。パッケージレベル変数の初期化のように、**引数が定数で、失敗するならそれはバグ** という場面で使います。

```go
var validID = regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)
```

> ⚠️ **ライブラリは panic で失敗を伝えてはいけません**。呼び出し側に処理させるべき失敗は、`error` で返します。

## 7.6 recover

`recover` を `defer` された関数の中で呼ぶと、panic を止めて、`panic` に渡された値を受け取れます。panic 中でなければ `nil` を返します。

### 1 つの処理の panic でプログラム全体を落とさない

サーバーでは、1 件のリクエスト処理で panic が起きても、他のリクエストに影響しないようにします。Effective Go の例を元にしています。

```go
package main

import (
	"fmt"
	"log"
	"os"
)

func safelyDo(work func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("work failed:", err)
		}
	}()
	work()
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	safelyDo(func() { fmt.Println("job 1 ok") })
	safelyDo(func() {
		var m map[string]int
		m["x"] = 1 // nil マップへの書き込みで panic
	})
	safelyDo(func() { fmt.Println("job 3 ok") })
}
```

```text
job 1 ok
work failed: assignment to entry in nil map
job 3 ok
```

`net/http` のサーバーも、ハンドラの panic を内部で recover しています。第15章では、panic を 500 エラーに変換するミドルウェアを書きます。

### パッケージの内部で panic を error に変換する

再帰が深いパーサーなどでは、エラーが起きるたびに何段も `return err` するのは大変です。そこで、**パッケージの内部では panic でエラーを伝え、公開関数の境界で recover して error に戻す** という手法が使われます。Go の `regexp` パッケージや `encoding/json` の内部でも使われています。

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// parseError は、このパッケージの内部だけで使う panic 用の型
type parseError struct{ msg string }

type parser struct{ tokens []string }

func (p *parser) fail(format string, args ...any) {
	panic(parseError{fmt.Sprintf(format, args...)})
}

func (p *parser) number() int {
	if len(p.tokens) == 0 {
		p.fail("unexpected end of input")
	}
	n, err := strconv.Atoi(p.tokens[0])
	if err != nil {
		p.fail("not a number: %q", p.tokens[0])
	}
	p.tokens = p.tokens[1:]
	return n
}

// Sum は "1 + 2 + 3" のような式を計算する公開関数
func Sum(expr string) (result int, err error) {
	defer func() {
		if e := recover(); e != nil {
			pe, ok := e.(parseError)
			if !ok {
				panic(e) // 想定外の panic (本物のバグ) は握りつぶさずに、もう一度 panic させる
			}
			// 名前付き戻り値を書き換えて error にする。
			// result には途中までの値が残っているので、ゼロに戻しておく
			result, err = 0, fmt.Errorf("parse %q: %s", expr, pe.msg)
		}
	}()
	p := &parser{tokens: strings.Fields(expr)}
	result = p.number()
	for len(p.tokens) > 0 {
		if p.tokens[0] != "+" {
			p.fail("expected '+', got %q", p.tokens[0])
		}
		p.tokens = p.tokens[1:]
		result += p.number()
	}
	return result, nil
}

func main() {
	fmt.Println(Sum("1 + 2 + 3"))
	fmt.Println(Sum("1 + x"))
	fmt.Println(Sum("1 +"))
	fmt.Println(Sum("1 * 2"))
}
```

```text
6 <nil>
0 parse "1 + x": not a number: "x"
0 parse "1 +": unexpected end of input
0 parse "1 * 2": expected '+', got "*"
```

ポイントは次の 3 つです。

- **panic がパッケージの外に漏れない** ようにします。公開関数は普通に `error` を返します。
- 型アサーションで **自分が起こした panic だけ** を error に変換します。それ以外の panic は `panic(e)` で再び起こします。
- 名前付き戻り値と `defer` を組み合わせて、関数の戻り値を書き換えています (第3章)。panic した時点の `result` には途中まで計算した値が残っているので、0 に戻しています。この 1 行を消すと、`Sum("1 + x")` が `1` とエラーを同時に返してしまいます。

## 7.7 📦 OSS ではこう書かれている

- **Kubernetes**: `k8s.io/apimachinery/pkg/api/errors` に `IsNotFound(err)` や `IsConflict(err)` のような判定関数を用意しています。呼び出し側はエラー型を直接扱わずに済みます。
- **GitHub CLI (cli/cli)**: `cmdutil.SilentError` や `cmdutil.FlagErrorf` のように、「どう表示するか」を表すエラー型を定義しています。最上位の `ghcmd.Main()` が、それらをもとに終了コードとメッセージをまとめて決めます。
- **golangci-lint の `errorlint`**: `err == ErrX` や `err.(*T)` のように、ラップを考慮しない比較を検出します。本リポジトリの `.golangci.yml` でも有効にしています。

## ✍️ 演習: エラーを返す Sqrt

[`exercises/errsqrt`](../exercises/errsqrt/exercise.go) を実装してください。

1. `type ErrNegativeSqrt float64` に `Error() string` メソッドを実装し、`"cannot Sqrt negative number: -2"` のようなメッセージを返します。
2. `Sqrt(x float64) (float64, error)` は、x が負の数なら `ErrNegativeSqrt(x)` を返します。

```bash
cd exercises
go test ./errsqrt/
```

**考えてみよう**: `Error()` メソッドの中で `fmt.Sprint(e)` と書くと、プログラムが無限ループになります。なぜでしょうか。ヒント: 第5章の Stringer の注意点と同じ理由です。
