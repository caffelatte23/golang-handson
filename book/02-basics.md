# 第2章 基本 — パッケージ・関数・変数・型・定数

この章から第8章までで、Go の言語仕様をひととおり学びます。「Tour of Go」と「Effective Go」の内容をすべて含め、現場でよく使う書き方や落とし穴も補っています。

`package main` と `func main()` を持つサンプルはそのまま実行できます。手元で `go run .` してもよいですし、[Go Playground](https://go.dev/play/) に貼り付けても試せます。サンプルの直後に載せた出力は、実際に実行して確かめたものです。

## 2.1 最初のプログラムを読む

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Gopher!")
}
```

```text
Hello, Gopher!
```

| 行 | 意味 |
| --- | --- |
| `package main` | このファイルが属するパッケージ。**実行ファイルになるのは `main` パッケージだけ** |
| `import "fmt"` | 標準ライブラリの `fmt` (フォーマット付き入出力) を使う宣言 |
| `func main()` | プログラムの開始地点。引数も戻り値もない |

## 2.2 パッケージと import

Go のプログラムはすべて **パッケージ** でできています。パッケージ名は、import パスの最後の要素と同じにするのが決まりです。たとえば `"math/rand"` のパッケージ名は `rand` です。

複数の import は、括弧でまとめて書く **グループ化 import** が一般的です。

```go
package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println(math.Sqrt(16))
	fmt.Println(strings.ToUpper("gopher"))
}
```

```text
4
GOPHER
```

- 1 つのディレクトリには 1 つのパッケージだけを置きます。
- **使っていない import はコンパイルエラー** になります。依存を整理し忘れないための仕組みです。
- `goimports` (第12章) を使うと、保存時に import を自動で追加・削除してくれます。
- 標準ライブラリ、空行、外部パッケージの順にグループを分けるのが慣習です。

## 2.3 エクスポートされる名前

**大文字で始まる名前だけがパッケージの外から参照できます (exported)**。`public` や `private` のようなキーワードはありません。

```go
math.Pi // OK: 大文字で始まるので公開されている
math.pi // コンパイルエラー: undefined: math.pi (小文字で始まる名前は外から見えない)
```

この規則は関数・型・変数・定数・構造体のフィールド・メソッドのすべてに当てはまります。

```go
type User struct {
	Name     string // 公開フィールド (JSON 変換などで外部パッケージから見える)
	password string // 非公開フィールド
}
```

## 2.4 関数

```go
package main

import "fmt"

// 引数の型は名前の後ろに書く
func add(x int, y int) int {
	return x + y
}

// 同じ型が続くときは最後以外を省略できる
func mul(x, y int) int {
	return x * y
}

// 複数の値を返せる
func swap(a, b string) (string, string) {
	return b, a
}

func main() {
	fmt.Println(add(42, 13), mul(6, 7))
	a, b := swap("hello", "world")
	fmt.Println(a, b)
}
```

```text
55 42
world hello
```

### なぜ型を後ろに書くのか

C の `int (*fp)(int (*)(int, int), int)` のような宣言は、内側から外側へ読み解く必要があります。Go の `fp func(func(int, int) int, int) int` は、左から右へそのまま読めます (Go Blog: [Go's Declaration Syntax](https://go.dev/blog/declaration-syntax))。

### 名前付き戻り値

戻り値に名前を付けると、関数の先頭で **ゼロ値に初期化された変数** として扱われます。

```go
package main

import "fmt"

// sum を 2 つに分ける
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return // naked return: 名前付き戻り値の現在の値を返す
}

func main() {
	fmt.Println(split(17))
}
```

```text
7 10
```

名前付き戻り値は、**戻り値の意味を説明するドキュメント** として役立ちます。

```go
func (f *File) Read(b []byte) (n int, err error)
func nextInt(b []byte, pos int) (value, nextPos int)
```

ただし、引数なしの `return` (naked return) は長い関数では読みにくくなります。**短い関数だけ** で使いましょう。名前付き戻り値は、`defer` の中で戻り値を書き換えるときにも使います (第3章)。

## 2.5 変数

```go
package main

import "fmt"

var debug bool                     // パッケージレベル変数 (ゼロ値 false)
var host, port = "localhost", 8080 // 初期値があれば型を省略できる

func main() {
	var i int    // ゼロ値 0
	var s = "go" // 型推論で string
	n := 10      // 短縮変数宣言 (関数の中だけで使える)
	x, y := 1.5, true

	fmt.Println(debug, host, port, i, s, n, x, y)
}
```

```text
false localhost 8080 0 go 10 1.5 true
```

- 関数の中では `:=` を使うのが普通です。
- 関数の外 (パッケージレベル) では `:=` は使えず、`var` が必要です。
- 宣言したまま **使っていないローカル変数はコンパイルエラー** になります。

### `:=` の再宣言ルール

`:=` の左辺に **新しい変数が 1 つ以上** あれば、同じスコープの既存の変数には単に代入されます。これにより `err` を使い回せます。

```go
f, err := os.Open(name)   // f と err を宣言
// ...
d, err := f.Stat()        // d は新規、err は既存の変数への代入
```

ただし、**内側のスコープで `:=` を使うと、外側と同じ名前の新しい変数ができてしまいます (シャドーイング)**。よくあるバグなので気を付けましょう。

```go
package main

import "fmt"

func main() {
	x := 1
	if true {
		x := 2 // 新しい x (外側の x とは別物)
		x++
	}
	fmt.Println(x)
}
```

```text
1
```

## 2.6 基本型

```text
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // uint8 の別名
rune // int32 の別名。Unicode のコードポイントを表す

float32 float64

complex64 complex128
```

`int`、`uint`、`uintptr` の大きさは、64 ビット環境では 64 ビットです。**整数は特に理由がなければ `int` を使います**。

```go
package main

import (
	"fmt"
	"math"
	"math/cmplx"
)

var (
	ToBe   bool       = false
	MaxInt uint64     = 1<<64 - 1
	z      complex128 = cmplx.Sqrt(-5 + 12i)
	r      rune       = 'G'
	f      float64    = math.Pi
)

func main() {
	fmt.Printf("%T: %v\n", ToBe, ToBe)
	fmt.Printf("%T: %v\n", MaxInt, MaxInt)
	fmt.Printf("%T: %v\n", z, z)
	fmt.Printf("%T: %v (%c)\n", r, r, r)
	fmt.Printf("%T: %.4f\n", f, f)
}
```

```text
bool: false
uint64: 18446744073709551615
complex128: (2+3i)
int32: 71 (G)
float64: 3.1416
```

`%T` は型を、`%v` は値を表示する書式指定子です (詳しくは第4章)。

### 文字列

- `string` は **変更できない (immutable) バイト列** で、中身は通常 UTF-8 です。
- `len(s)` は **バイト数** を返します。文字数ではありません。
- `s[i]` は i 番目の **バイト** (`byte`) です。
- `for range` で回すと、UTF-8 をデコードしながら **rune (文字) 単位** で取り出せます。

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "Go言語"
	fmt.Println(len(s), utf8.RuneCountInString(s))
	for i, r := range s {
		fmt.Printf("%d:%c ", i, r) // i はバイト位置
	}
	fmt.Println()

	raw := `C:\path\no\escape` // バッククォートは生文字列 (エスケープなし・改行可)
	fmt.Println(raw)
}
```

```text
8 4
0:G 1:o 2:言 5:語 
C:\path\no\escape
```

## 2.7 ゼロ値

明示的に初期化しなかった変数には、型ごとの **ゼロ値** が入ります。Go には「未初期化の変数」がありません。

```go
package main

import "fmt"

func main() {
	var i int
	var f float64
	var b bool
	var s string
	var p *int
	var sl []int
	var m map[string]int
	fmt.Printf("%v %v %v %q %v %v %v\n", i, f, b, s, p, sl, m)
	fmt.Println(sl == nil, m == nil)
}
```

```text
0 0 false "" <nil> [] map[]
true true
```

| 型 | ゼロ値 |
| --- | --- |
| 数値型 | `0` |
| `bool` | `false` |
| `string` | `""` (空文字列) |
| ポインタ、スライス、マップ、チャネル、関数、インターフェース | `nil` |
| 配列、構造体 | 各要素・各フィールドがゼロ値 |

> 💡 **ゼロ値を役立つように設計する** (*Make the zero value useful*)
>
> `sync.Mutex` は宣言しただけでロックが解除された状態として使えます。`bytes.Buffer` も宣言しただけで空のバッファとして使えます。自分で型を作るときも、**`var x T` のまま意味のある動作をする** ように設計すると、使う側の初期化の手間がなくなります。

## 2.8 型変換

Go には **暗黙の型変換がありません**。`int` と `float64` を混ぜて計算するときも、`T(v)` の形で明示的に変換します。

```go
package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	x, y := 3, 4
	f := math.Sqrt(float64(x*x + y*y)) // int → float64
	z := uint(f)                       // float64 → uint (小数点以下は切り捨て)
	fmt.Println(x, y, z)

	// 数値と文字列の変換は strconv を使う
	s := strconv.Itoa(42)
	n, err := strconv.Atoi("123")
	fmt.Println(s, n, err)

	fmt.Println(string(rune(65))) // rune → 文字。string(65) は "A" になるが vet が警告する
}
```

```text
3 4 5
42 123 <nil>
A
```

## 2.9 型推論

右辺に型が付いている値なら、左辺の型は同じになります。型の付いていない数値定数の場合は、書き方で型が決まります。

```go
i := 42           // int
f := 3.142        // float64
g := 0.867 + 0.5i // complex128
```

## 2.10 定数

```go
package main

import "fmt"

const Pi = 3.14

const (
	Big   = 1 << 100  // 型なし定数は任意精度で計算される
	Small = Big >> 99 // 2
)

func needInt(x int) int           { return x*10 + 1 }
func needFloat(x float64) float64 { return x * 0.1 }

func main() {
	const World = "世界"
	fmt.Println("Hello", World, Pi)
	fmt.Println(needInt(Small))
	fmt.Println(needFloat(Small))
	fmt.Println(needFloat(Big))
	// fmt.Println(needInt(Big)) // コンパイルエラー: 1<<100 は int に収まらない
}
```

```text
Hello 世界 3.14
21
0.2
1.2676506002282295e+29
```

- 定数は `const` で宣言します。`:=` は使えません。
- 定数にできるのは、文字・文字列・真偽値・数値だけです。値は **コンパイル時に決まる式** でなければなりません。`math.Sin(1)` のような関数呼び出しは使えません。
- **型を付けない数値定数は高い精度を持ちます**。使われた場所の型に合わせて変換されます。

### iota による連番

`const` ブロックの中では、`iota` が 0, 1, 2, ... と 1 つずつ増えます。式を省略すると、直前の式がそのまま繰り返されます。

```go
package main

import "fmt"

type Weekday int

const (
	Sunday  Weekday = iota // 0
	Monday                 // 1
	Tuesday                // 2
)

type ByteSize float64

const (
	_           = iota             // 0 は空白識別子で捨てる
	KB ByteSize = 1 << (10 * iota) // 1 << 10
	MB                             // 1 << 20
	GB                             // 1 << 30
)

func main() {
	fmt.Println(Sunday, Monday, Tuesday)
	fmt.Println(KB, MB, GB)
}
```

```text
0 1 2
1024 1.048576e+06 1.073741824e+09
```

この `ByteSize` に `String()` メソッドを付けて "1.00MB" のように表示する例は、第5章で扱います。

## 2.11 初期化と init 関数

パッケージレベルの変数は、**依存関係を考慮して** 初期化されます。規則は「まだ初期化されていない変数に依存していないもののうち、宣言順で最も前にあるものから初期化する」です。

```go
var (
	a = c + b // 3 番目: b と c の両方が初期化されるまで待つ
	b = f()   // 1 番目: 何にも依存していない (f が a や c を参照しない場合)
	c = 1     // 2 番目
)
```

ソースに書いた順ではないことに注意してください。ただし、依存関係のない変数どうしは宣言順に初期化されます。

各ファイルには `init()` 関数を置けます。パッケージ内の変数がすべて初期化されたあと、`main` より前に自動で呼ばれます。

```go
package main

import "fmt"

var config = loadDefaults()

func loadDefaults() map[string]string {
	fmt.Println("1. 変数の初期化")
	return map[string]string{"env": "dev"}
}

func init() {
	fmt.Println("2. init")
	config["env"] = "test"
}

func main() {
	fmt.Println("3. main:", config["env"])
}
```

```text
1. 変数の初期化
2. init
3. main: test
```

> ⚠️ `init` は便利ですが、**暗黙に実行される処理** です。テストしにくく、副作用が見えにくくなります。宣言だけでは表せない初期化 (計算したテーブルの準備など) に限り、ネットワーク接続やファイル読み込みのような重い処理は `main` で明示的に行いましょう。

## 2.12 セミコロンとブレースの位置

Go の文法上は文の終わりにセミコロンが必要です。ただし **字句解析器が行末に自動で挿入する** ので、ソースに書くことはほとんどありません。自動挿入の規則は「行の最後のトークンが識別子・リテラル・`)`・`}`・`return` などなら、その後ろにセミコロンを入れる」というものです。

その結果、**開きブレース `{` は必ず同じ行に書く** 必要があります。

```go
if x > 0 {   // ✅
}

if x > 0     // ❌ ここにセミコロンが挿入されて構文エラー
{
}
```

ブレースの位置で議論になることはありません。`gofmt` が全員のコードを同じ形に整えるからです (第9章)。

## 2.13 名前の付け方 (概要)

Go の命名には強い慣習があります。詳しくは第9章で扱い、ここでは要点だけ挙げます。

- 複数の単語は **MixedCaps / mixedCaps** でつなぎます。アンダースコアは使いません (`maxLength`、`MaxLength`。`max_length` は使わない)。
- `URL`、`ID`、`HTTP` のような **頭字語は大文字か小文字でそろえます** (`userID`、`ServeHTTP`、`urlPath`)。
- **スコープが狭いほど名前を短く** します。ループ変数は `i`、Reader は `r`、レシーバは `s` のように 1〜2 文字が普通です。

## ✍️ 練習

1. `int8` に 127 を入れて 1 を足すとどうなるか確かめよう (オーバーフローの挙動)。
2. `const` で `KB` から `TB` までを定義し、`fmt.Printf("%.0f\n", TB)` で表示しよう。
3. `x := 1; x := 2` が同じスコープではコンパイルエラーになるのに、`x, y := 1, 2` の後の `x, z := 3, 4` が通る理由を説明しよう。
