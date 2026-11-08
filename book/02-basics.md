# 第2章 基本文法

この章では Go の文法を「他の言語との違い」を意識しながら駆け足で確認します。コード片はすべて [Go Playground](https://go.dev/play/) に貼り付けて試せます。

## 2.1 パッケージと import

```go
package main // 実行ファイルになるのは package main + func main()

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra" // 標準ライブラリ → 空行 → 外部パッケージ の順が慣習
)
```

- 1 ディレクトリ = 1 パッケージ。
- **大文字で始まる識別子だけが外部に公開 (exported)** されます。`public` / `private` キーワードはありません。

```go
func Parse() {} // 他パッケージから呼べる
func parse() {} // パッケージ内だけ
```

- 未使用の import / 変数は **コンパイルエラー**。

## 2.2 変数・定数・型

```go
var a int          // ゼロ値 0 で初期化
var b = "hello"    // 型推論
c := 3.14          // 関数内だけで使える短縮宣言 (最もよく使う)
x, y := 1, 2

const Pi = 3.14159
```

### ゼロ値

Go の変数は必ず **ゼロ値** で初期化されます。未初期化変数という概念がありません。

| 型 | ゼロ値 |
| --- | --- |
| 数値 | `0` |
| `string` | `""` |
| `bool` | `false` |
| ポインタ, スライス, マップ, チャネル, 関数, インターフェース | `nil` |
| 構造体 | 各フィールドがゼロ値 |

> 📦 **OSS ではこう書かれている**: `sync.Mutex` や `bytes.Buffer` は「ゼロ値のまま使える」ように設計されています。自作の型でも「`var x T` でそのまま動く」設計を目指すのが Go 流です (Go Proverbs: *Make the zero value useful*)。本書のハンズオン `todo.List` もゼロ値で使えるようにしています。

### 列挙型の代わりに `iota`

```go
type Status int

const (
	StatusTodo Status = iota // 0
	StatusDoing              // 1
	StatusDone               // 2
)

// String() を実装すると fmt で表示したときに使われる
func (s Status) String() string {
	switch s {
	case StatusTodo:
		return "todo"
	case StatusDoing:
		return "doing"
	case StatusDone:
		return "done"
	}
	return fmt.Sprintf("Status(%d)", int(s))
}
```

> 💡 大量の定数に `String()` を書くのは面倒なので、OSS では `go generate` + [`stringer`](https://pkg.go.dev/golang.org/x/tools/cmd/stringer) で自動生成するのが定番です。

## 2.3 制御構文

### if — 初期化文付き

```go
if err := doSomething(); err != nil {
	return err
}
// err はここではスコープ外
```

### for — ループはこれだけ

```go
// 構文の一覧 (実際には宣言した変数を使わないとコンパイルエラーになります)
for i := 0; i < 10; i++ {}      // C 風
for cond {}                      // while 相当
for {}                           // 無限ループ
for i, v := range []string{"a"} {} // スライス
for k, v := range map[string]int{} {} // マップ (順序は不定!)
for i := range 10 {}             // 整数 range (Go 1.22+)
```

> ⚠️ Go 1.22 で **ループ変数がイテレーションごとに新しく作られる** よう仕様が変わりました。古い記事にある `v := v` のようなコピーは不要です (lint の `copyloopvar` が検出します)。

### switch — break 不要

```go
switch os := runtime.GOOS; os {
case "darwin", "linux":
	fmt.Println("unix")
default:
	fmt.Println(os)
}

// 条件なし switch は if-else の連鎖をきれいに書ける
switch {
case score >= 80:
	grade = "A"
case score >= 60:
	grade = "B"
default:
	grade = "C"
}
```

## 2.4 関数

```go
// 複数の戻り値。最後を error にするのが Go の約束事
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// 可変長引数
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 関数は値。クロージャも書ける
counter := func() func() int {
	n := 0
	return func() int { n++; return n }
}()
```

### defer

関数を抜けるときに実行されます。後片付けに使います。

```go
f, err := os.Open("data.txt")
if err != nil {
	return err
}
defer f.Close() // 開いた直後に閉じる処理を書くのがイディオム
```

## 2.5 スライスとマップ

```go
s := []int{1, 2, 3}
s = append(s, 4)          // append の戻り値を必ず代入する
sub := s[1:3]             // [2 3] (元の配列を共有している点に注意)
buf := make([]byte, 0, 1024) // 長さ 0, 容量 1024

m := map[string]int{"apple": 100}
m["banana"] = 200
v, ok := m["cherry"] // ok は存在確認 (comma ok イディオム)
delete(m, "apple")
```

Go 1.21 以降は標準の `slices` / `maps` パッケージが便利です。

```go
import ("slices"; "maps")

slices.Sort(s)
idx := slices.Index(s, 3)
has := slices.Contains(s, 4)
keys := slices.Sorted(maps.Keys(m)) // Go 1.23+ (イテレータ)
```

## 2.6 文字列

- `string` はイミュータブルなバイト列 (UTF-8)。
- `len("日本")` は **6** (バイト数)。文字 (rune) 単位で扱うなら `[]rune(s)` や `for _, r := range s`。
- 連結を繰り返すなら `strings.Builder`。

```go
var sb strings.Builder
for i := range 3 {
	fmt.Fprintf(&sb, "line %d\n", i)
}
fmt.Print(sb.String())
```

## 2.7 構造体とメソッド

```go
type User struct {
	ID    int64  `json:"id"`              // 構造体タグ (JSON 変換などで使う)
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// 値レシーバ: コピーを受け取る (読み取り専用)
func (u User) Display() string { return u.Name + " <" + u.Email + ">" }

// ポインタレシーバ: 元の値を変更できる
func (u *User) Rename(name string) { u.Name = name }

u := User{ID: 1, Name: "Gopher"}
u.Rename("Gordon") // u は自動的に &u として渡される
```

**レシーバの選び方** ([Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#receiver-type) より要約):

- 値を変更する / 大きな構造体 / `sync.Mutex` を含む → ポインタレシーバ
- 迷ったらポインタレシーバ。同じ型のメソッドでは揃える

### コンストラクタ関数

Go にコンストラクタ構文はなく、`NewXxx` 関数を作るのが慣習です。

```go
func NewServer(addr string, opts ...Option) *Server { ... }
```

### 埋め込み (Embedding)

継承はありませんが、型を **埋め込む** ことでメソッドを昇格できます (委譲)。

```go
type Logger struct{}
func (Logger) Log(msg string) { fmt.Println(msg) }

type Service struct {
	Logger // 埋め込み
	name string
}

s := Service{name: "api"}
s.Log("started") // Logger.Log が呼べる
```

## 2.8 ポインタ

```go
x := 10
p := &x   // アドレス取得
*p = 20   // 間接参照
fmt.Println(x) // 20
```

- ポインタ演算はありません (安全)。
- `new(T)` は `T` のゼロ値へのポインタを返します。実際には `&T{...}` のほうがよく使われます。

## 2.9 Functional Options パターン

設定項目が多い構造体を作るとき、OSS で非常によく使われるパターンです。

```go
type Server struct {
	addr    string
	timeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
	return func(s *Server) { s.timeout = d }
}

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 30 * time.Second} // デフォルト値
	for _, opt := range opts {
		opt(s)
	}
	return s
}

srv := NewServer(":8080", WithTimeout(5*time.Second))
```

> 📦 **OSS ではこう書かれている**: gRPC-Go の `grpc.NewServer(opts ...grpc.ServerOption)`、zap の `zap.New(core, zap.AddCaller())` などがこのパターンです。出典は Dave Cheney の記事 *Functional options for friendly APIs* (2014)。

## ✍️ 練習問題

1. 文字列中の単語の出現回数を数える `WordCount(s string) map[string]int` を書いてみよう (`strings.Fields` を使う)。
2. `Rectangle` と `Circle` 構造体を作り、それぞれに `Area() float64` メソッドを実装しよう (次章でインターフェースにします)。
