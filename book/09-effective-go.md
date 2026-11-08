# 第9章 Go らしいコードの書き方 (Effective Go)

ここまでで言語機能はひととおり学びました。この章では、**Go コミュニティで共有されている書き方の約束事** をまとめます。内容は Effective Go と Go Code Review Comments を元にしています。

これらの約束事に沿って書くと、他の Go プログラマが読みやすくなり、OSS の慣習にもなじみます。

## 9.1 フォーマット — gofmt に任せる

Go では **コードの見た目を人が決めません**。`gofmt` (または `go fmt`) がすべて決めます。

```bash
gofmt -w .      # ファイルを書き換える
gofmt -l .      # 整形が必要なファイルの一覧を表示する (CI で使う)
gofmt -d .      # 差分を表示する
```

- インデントには **タブ** を使います。gofmt がそうするので、議論する必要はありません。
- **1 行の長さに制限はありません**。長すぎると感じたら、改行してタブ 1 つでインデントします。
- 括弧は少なめです。`if`、`for`、`switch` の条件は括弧で囲みません。演算子の間隔は優先順位を表すように調整されます (`x<<8 + y<<16`)。
- 構造体のフィールドやコメントの縦の位置もそろえてくれます。

> *Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.*
> (gofmt のスタイルは誰の好みでもないが、gofmt は誰もが好きだ) — Go Proverbs

保存するたびに `goimports` (gofmt の機能に加えて import の自動整理もする) が実行されるように、エディタを設定しておきましょう。

## 9.2 コメントとドキュメント

Go では、**宣言の直前に書いたコメントがそのままドキュメントになります** (doc comment)。`go doc` コマンドや [pkg.go.dev](https://pkg.go.dev/) は、このコメントからドキュメントを生成します。

```go
// Package lru はジェネリクスを使ったスレッドセーフな LRU キャッシュです。
//
// 最も長く使われていない要素から追い出します。
package lru

// Cache は容量が固定の LRU キャッシュです。
// ゼロ値は使えないので、New で作成してください。
type Cache[K comparable, V any] struct { /* ... */
}

// New は容量 capacity のキャッシュを作成します。
// capacity が 1 未満のときは panic します。
func New[K comparable, V any](capacity int) *Cache[K, V] { /* ... */ }
```

**書き方の規則**

- **公開しているすべての名前** (大文字で始まる名前) に doc comment を書きます。
- **最初の文は、その名前で始まる完全な文** にします (`// New は...`、英語なら `// New creates ...`)。`go doc` の一覧には最初の文が表示されるので、それだけで用途が伝わるようにします。
- パッケージの doc comment は `// Package xxx ...` で始め、パッケージ内のどれか 1 つのファイルに書きます。長くなるときは `doc.go` という専用のファイルを作るのが慣習です。
- 非推奨にした API には、`// Deprecated: 代わりに NewV2 を使ってください。` という段落を加えます。エディタや linter が警告を出してくれます。

### doc comment の書式 (Go 1.19+)

```go
// Package json implements encoding and decoding of JSON.
//
// # Encoding                        ← "# " で始まる行は見出しになる
//
// Marshal は値を JSON に変換します。詳しくは [Marshal] を参照してください。   ← [ ] でリンク
//
//   - リスト項目 1                   ← "-" で始まる行は箇条書きになる
//   - リスト項目 2
//
// 例:
//
//	b, err := json.Marshal(v)        ← タブでインデントした行はコードブロックになる
package json
```

```bash
go doc fmt.Println         # 関数のドキュメント
go doc -all ./internal/... # パッケージ全体
```

第11章で扱う **Example テスト** は、ドキュメントに実行可能なサンプルとして表示されます。

## 9.3 命名

> 名前は、読み手が最初に目にするドキュメントです。

### パッケージ名

- **短く、すべて小文字で、1 単語** にします。アンダースコアや mixedCaps は使いません (`bufio`、`strconv`、`httptest`)。
- パッケージ名は **import パスの最後の要素** と同じにします (`encoding/base64` のパッケージ名は `base64`)。
- **呼び出し側では、名前の前にパッケージ名が付きます**。そのため、パッケージ名と重複する言葉は名前に入れません。

| ❌ | ✅ | 理由 |
| --- | --- | --- |
| `bufio.BufReader` | `bufio.Reader` | `bufio.` が付くので `Buf` は重複している |
| `ring.NewRing` | `ring.New` | パッケージの主要な型を作る関数は `New` だけでよい |
| `http.HTTPServer` | `http.Server` | |
| `user.UserService` | `user.Service` | |

- `util`、`common`、`misc`、`helpers`、`base` のように **中身を表さないパッケージ名は避けます**。何でも入る場所になってしまい、依存関係も複雑になります。

### ゲッターとセッター

Go では、ゲッターの名前に **`Get` を付けません**。

```go
owner := obj.Owner()       // ✅ GetOwner() とはしない
if owner != user {
	obj.SetOwner(user)     // セッターには Set を付ける
}
```

### インターフェース名

メソッドが 1 つのインターフェースは、**メソッド名に -er を付けた名前** にします (`Reader`、`Writer`、`Formatter`、`CloseNotifier`)。

また、`Read`、`Write`、`Close`、`Flush`、`String` のような **よく知られた名前のメソッドは、よく知られた意味でだけ使います**。たとえば、文字列に変換するメソッドの名前は `ToString` ではなく `String` にします。

### MixedCaps と頭字語

- 複数の単語は `MixedCaps` (公開する名前) か `mixedCaps` (公開しない名前) でつなぎます。アンダースコアは使いません。
- `URL`、`HTTP`、`ID`、`API` のような **頭字語は、大文字か小文字のどちらかにそろえます**。

| ❌ | ✅ |
| --- | --- |
| `Url`, `HttpServer`, `userId` | `URL`, `HTTPServer`, `userID` |
| `xmlHttpRequest` | `xmlHTTPRequest` |

### 名前の長さはスコープに比例させる

```go
for i, v := range items { ... }       // ループ変数は 1 文字でよい
func (c *Client) Do(r *Request) ...   // レシーバや引数も短くてよい
var defaultHTTPTimeout = 30 * time.Second // パッケージ全体から見える名前は説明的にする
```

### その他の命名の慣習

| 対象 | 慣習 | 例 |
| --- | --- | --- |
| エラー変数 | `Err` で始める | `ErrNotFound`, `io.EOF` (歴史的な例外) |
| エラー型 | `Error` で終える | `PathError`, `SyntaxError` |
| コンストラクタ | `New` / `NewXxx` | `bufio.NewReader`, `list.New` |
| レシーバ | 型名の 1〜2 文字の略。`this` や `self` は使わない | `func (s *Server)` |
| bool を返す関数 | `Is`、`Has`、`Can` で始める | `IsNotExist`, `HasPrefix` |

## 9.4 空白識別子 `_`

`_` には何でも代入でき、代入した値は捨てられます。

```go
// 1. 複数の戻り値のうち、要らないものを捨てる
if _, err := os.Stat(path); os.IsNotExist(err) {
	fmt.Printf("%s does not exist\n", path)
}

// ❌ エラーを捨ててはいけない (失敗したときに何が起きたか分からなくなる)
fi, _ := os.Stat(path)
if fi.IsDir() { ... } // path が存在しなければ fi は nil なので panic する

// 2. 副作用 (init 関数) のためだけに import する
import _ "net/http/pprof" // /debug/pprof/ ハンドラを登録する
import _ "image/png"      // image.Decode が PNG を読めるようになる

// 3. インターフェースを満たしているかをコンパイル時に確かめる
var _ http.Handler = (*Server)(nil)

// 4. 開発中に、使っていない import や変数のエラーを一時的に黙らせる (コミット前には消す)
var _ = fmt.Printf
```

## 9.5 設計パターン: Functional Options

設定項目が多い型のコンストラクタでは、OSS で **Functional Options** パターンがよく使われます。必須の引数は普通に受け取り、省略可能な設定は「設定を変更する関数」として可変長引数で受け取ります。

```go
package main

import (
	"fmt"
	"time"
)

type Server struct {
	addr    string
	timeout time.Duration
	maxConn int
}

// Option は Server の設定を変更する関数
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
	return func(s *Server) { s.timeout = d }
}

func WithMaxConn(n int) Option {
	return func(s *Server) { s.maxConn = n }
}

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 30 * time.Second, maxConn: 100} // デフォルト値
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func main() {
	fmt.Printf("%+v\n", *NewServer(":8080"))
	fmt.Printf("%+v\n", *NewServer(":8080", WithTimeout(5*time.Second), WithMaxConn(10)))
}
```

```text
{addr::8080 timeout:30000000000 maxConn:100}
{addr::8080 timeout:5000000000 maxConn:10}
```

`timeout` が `30s` ではなくナノ秒の数値で表示されている点に注目してください。`fmt` は **非公開フィールドの値の `String()` メソッドを呼べない** ため、`time.Duration` の元の型である int64 として表示されます。フィールドを公開する (`Timeout`) か、`s.timeout` を直接 `Println` すれば `30s` と表示されます。

- デフォルト値が 1 か所にまとまります。
- 設定項目を後から追加しても、**既存の呼び出し側のコードは変わりません** (後方互換性が保たれます)。
- gRPC-Go の `grpc.NewServer(opts ...grpc.ServerOption)` や zap の `zap.New(core, zap.AddCaller())` などがこのパターンを使っています。考案したのは Dave Cheney の記事 *Functional options for friendly APIs* (2014) です。

設定項目が少ないなら、`Config` 構造体を引数に取るほうが単純です (`http.Server{Addr: ..., ReadTimeout: ...}` はその例です)。

## 9.6 総まとめ: 小さな Web サーバー

Effective Go の最後には、ここまでの要素を使った小さな Web サーバーの例があります。ここでは同じ考え方で、**フォームに入力した文字列の文字数と単語数を表示するサーバー** を作ります。

<!-- nocheck: サーバーとして起動し続けるため -->
```go
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
)

var addr = flag.String("addr", ":1718", "http service address")

// html/template は値を自動的にエスケープするので、XSS を防げる
var page = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html>
<head><title>Word Counter</title></head>
<body>
{{if .}}
<p>文字数: {{.Chars}} / 単語数: {{.Words}}</p>
<pre>{{.Text}}</pre>
{{end}}
<form action="/" method="GET">
  <textarea name="s" rows="4" cols="60"></textarea>
  <button type="submit">数える</button>
</form>
</body>
</html>
`))

type result struct {
	Text  string
	Chars int
	Words int
}

func count(w http.ResponseWriter, r *http.Request) {
	var res *result
	if s := r.FormValue("s"); s != "" {
		res = &result{Text: s, Chars: utf8.RuneCountInString(s), Words: len(strings.Fields(s))}
	}
	if err := page.Execute(w, res); err != nil {
		log.Println(err)
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("GET /", count)
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

```bash
go run . && open http://localhost:1718/   # ブラウザで開く
```

この短いプログラムでは、次の要素を使っています。

- `flag`: コマンドライン引数を受け取ります。
- `template.Must`: パッケージレベル変数の初期化に失敗したら panic します (第7章)。
- `html/template`: 埋め込む値をエスケープして、安全な HTML を出力します。
- `http.HandleFunc`: 普通の関数を、Handler インターフェースを満たす値に変換して登録します (第5章)。
- ポインタの nil を使った「結果があるときだけ表示する」処理。

> ⚠️ これは学習用の最小構成です。本番で使うサーバーでは、タイムアウトの設定、グレースフルシャットダウン、構造化ログなどが必要です。第15章のハンズオンで、それらを含めて作ります。

## 9.7 Go Proverbs

Rob Pike が GopherFest 2015 で示した [Go Proverbs](https://go-proverbs.github.io/) は、Go の設計思想を短い言葉にまとめたものです。

| 格言 | 意味 | 本書の該当箇所 |
| --- | --- | --- |
| Don't communicate by sharing memory, share memory by communicating. | 共有メモリで通信せず、通信でメモリを共有する | 第8章 |
| Concurrency is not parallelism. | 並行と並列は違う | 第8章 |
| Channels orchestrate; mutexes serialize. | channel は協調させ、mutex は直列化する | 第8章 |
| The bigger the interface, the weaker the abstraction. | インターフェースは大きいほど抽象として弱くなる | 第5章 |
| Make the zero value useful. | ゼロ値を役に立つようにする | 第2章 |
| `interface{}` says nothing. | any は何も語らない (型の情報が失われる) | 第5章 |
| Gofmt's style is no one's favorite, yet gofmt is everyone's favorite. | gofmt の書式は誰の好みでもないが、誰もが gofmt を好む | 9.1 |
| A little copying is better than a little dependency. | 少しのコピーは少しの依存よりよい | 第13章 |
| Clear is better than clever. | 賢いコードより明快なコード | — |
| Errors are values. | エラーは値である | 第7章 |
| Don't just check errors, handle them gracefully. | エラーは確認するだけでなく、適切に処理する | 第7章 |
| Don't panic. | panic するな | 第7章 |
| Documentation is for users. | ドキュメントは使う人のために書く | 9.2 |

## 9.8 コードレビューのチェックリスト

[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) から、よく指摘される項目をまとめました。

- [ ] `gofmt` (`goimports`) で整形されている
- [ ] 公開しているすべての名前に、その名前で始まる doc comment がある
- [ ] エラーを `_` で捨てていない
- [ ] エラーメッセージは小文字で始まり、末尾に句点がない
- [ ] エラーは「処理する」か「文脈を付けて返す」のどちらか一方だけをしている
- [ ] `panic` を通常のエラー処理に使っていない
- [ ] 頭字語が `URL`、`ID` のようにそろっている
- [ ] ゲッターに `Get` を付けていない
- [ ] レシーバ名が短く一貫していて、`this` や `self` を使っていない
- [ ] インターフェースは使う側で定義され、小さい
- [ ] goroutine には終了する条件がある (リークしない)
- [ ] `context.Context` は第 1 引数で受け取り、構造体に保存していない
- [ ] パッケージ名に `util` や `common` を使っていない
- [ ] テストの失敗メッセージが `got X, want Y` の形で、何が起きたか分かる
- [ ] import は、標準ライブラリと外部パッケージのグループに分かれている

## 次の一歩

第1部はここで終わりです。第2部では、こうした約束事が **実際の OSS でどのように実践されているか** を、プロジェクトの構成、テスト、ツール、ライブラリ選定の観点から調べていきます。
