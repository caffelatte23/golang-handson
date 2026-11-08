# 第5章 メソッドとインターフェース

Go にはクラスも継承もありません。その代わりに **メソッド**、**インターフェース**、**埋め込み** の 3 つを組み合わせて抽象化を表現します。

## 5.1 メソッド

メソッドは **レシーバ** という特別な引数を持つ関数です。レシーバは `func` キーワードとメソッド名の間に書きます。

```go
package main

import (
	"fmt"
	"math"
)

type Vertex struct{ X, Y float64 }

// Abs は Vertex 型をレシーバに持つメソッド
func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// 同じ処理を普通の関数として書いた場合
func Abs(v Vertex) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func main() {
	v := Vertex{3, 4}
	fmt.Println(v.Abs(), Abs(v))
}
```

```text
5 5
```

### 構造体以外の型にもメソッドを定義できる

**同じパッケージで定義した型** であれば、どの型にもメソッドを付けられます。`int` や他のパッケージの型には直接付けられないので、新しい型を定義してから付けます。

```go
package main

import (
	"fmt"
	"math"
)

type MyFloat float64

func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func main() {
	f := MyFloat(-math.Sqrt2)
	fmt.Println(f.Abs())
}
```

```text
1.4142135623730951
```

次の例は Effective Go にあるもので、`ByteSize` 型に `String` メソッドを付けて、単位付きで表示されるようにしています。

```go
package main

import "fmt"

type ByteSize float64

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
)

func (b ByteSize) String() string {
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

func main() {
	fmt.Println(ByteSize(1e9), ByteSize(1500), ByteSize(12))
}
```

```text
953.67MB 1.46KB 12.00B
```

`Sprintf` に `b` そのものではなく `b/GB` (結果は `ByteSize` 型) を `%f` で渡している点に注意してください。`%f` は数値として書式化するので、`String()` は呼ばれません。もし `%v` で `b` を渡すと `String()` が再び呼ばれ、無限再帰になります。

## 5.2 ポインタレシーバ

レシーバの型を `*T` にすると、メソッドの中で **元の値を変更** できます。

```go
package main

import (
	"fmt"
	"math"
)

type Vertex struct{ X, Y float64 }

func (v Vertex) Abs() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y) }

func (v *Vertex) Scale(f float64) {
	v.X *= f
	v.Y *= f
}

func (v Vertex) ScaleCopy(f float64) {
	v.X *= f // コピーを変更しているだけなので、呼び出し元には影響しない
	v.Y *= f
}

func main() {
	v := Vertex{3, 4}
	v.ScaleCopy(10)
	fmt.Println(v)
	v.Scale(10) // v はアドレスを取れるので、Go が自動的に (&v).Scale(10) として呼ぶ
	fmt.Println(v, v.Abs())

	p := &Vertex{4, 3}
	fmt.Println(p.Abs()) // 逆も同じで、(*p).Abs() として呼ばれる
}
```

```text
{3 4}
{30 40} 50
5
```

### 呼び出しの自動変換と、インターフェースを満たすかどうかの規則

| | 値 `v` から呼ぶ | ポインタ `p` から呼ぶ |
| --- | --- | --- |
| 値レシーバのメソッド | ✅ | ✅ (`(*p).M()` に自動変換) |
| ポインタレシーバのメソッド | ✅ **ただし v のアドレスが取れる場合だけ** (`(&v).M()` に自動変換) | ✅ |

メソッドを直接呼ぶときは、上の表のように Go が自動で変換してくれます。しかし **インターフェースを満たすかどうか** を判定するときは、この自動変換は行われません。

- `T` 型のメソッドセット (その型が持つメソッドの集合) には、**値レシーバのメソッドだけ** が含まれます。
- `*T` 型のメソッドセットには、**値レシーバとポインタレシーバの両方のメソッド** が含まれます。

そのため、ポインタレシーバのメソッドでインターフェースを実装した場合、**そのインターフェースを満たすのは `*T` だけ** です (5.4 節で実際に確かめます)。

この規則があるのは、値をインターフェースに入れるとコピーされるからです。コピーに対してポインタレシーバのメソッドを呼んでも、元の値は変わりません。そのような変更が黙って失われるのを防いでいます。

### 値レシーバとポインタレシーバの選び方

[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#receiver-type) の指針をまとめると、次のとおりです。

- レシーバを **変更する** なら、ポインタレシーバにします。
- `sync.Mutex` など **コピーしてはいけないフィールド** を持つなら、ポインタレシーバにします。
- 構造体が **大きい** なら、コピーのコストを避けるためにポインタレシーバにします。
- `time.Time` のような小さな値型や、マップ・関数・チャネルを元にした型なら、値レシーバでかまいません。
- **1 つの型の中では、どちらかにそろえます**。迷ったらポインタレシーバにします。

レシーバの名前は、型名の頭文字を取った 1〜2 文字にするのが慣習です (`func (c *Client)`)。`this` や `self` は使いません。

## 5.3 インターフェース

**インターフェース型** は、メソッドのシグネチャの集まりを定義した型です。

```go
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }

func (r Rect) Area() float64        { return r.W * r.H }
func (r Rect) Perimeter() float64   { return 2 * (r.W + r.H) }
func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func describe(s Shape) {
	fmt.Printf("%-8s area=%6.2f perimeter=%6.2f\n", fmt.Sprint(s), s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	for _, s := range shapes {
		describe(s)
	}
}
```

```text
{3 4}    area= 12.00 perimeter= 14.00
{1}      area=  3.14 perimeter=  6.28
```

### インターフェースは宣言しなくても実装される

`Rect` には `implements Shape` のような宣言がありません。**必要なメソッドをすべて持っていれば、自動的にそのインターフェースを満たします**。これを構造的部分型 (structural typing) と呼びます。

この仕組みには次の利点があります。

- 実装するパッケージは、インターフェースを定義したパッケージを import する必要がありません。
- **既存の型にも、あとからインターフェースを当てはめられます**。標準ライブラリの `*os.File` は、作られた時点では存在しなかった自作のインターフェースも満たせます。
- そのため、インターフェースは **使う側のパッケージで、必要なメソッドだけを持つ小さな形で定義する** のが Go の流儀です。

### インターフェース値の中身は (具体的な値, 型) の組

インターフェース型の変数には、**具体的な値** と **その値の型** の 2 つが入っています。メソッドを呼ぶと、入っている型のメソッドが実行されます。

```go
package main

import "fmt"

type I interface{ M() }

type T struct{ S string }

func (t *T) M() {
	if t == nil { // nil レシーバでも呼ばれる。nil を正しく扱うのは実装側の責任
		fmt.Println("<nil>")
		return
	}
	fmt.Println(t.S)
}

func main() {
	var i I

	var t *T
	i = t // 中身の値は nil だが、型 (*T) は入っている
	fmt.Printf("(%v, %T)\n", i, i)
	i.M()
	fmt.Println(i == nil) // ← false になることに注意!

	i = &T{"hello"}
	fmt.Printf("(%v, %T)\n", i, i)
	i.M()
}
```

```text
(<nil>, *main.T)
<nil>
false
(&{hello}, *main.T)
hello
```

> ⚠️ **nil の入ったポインタを持つインターフェースは nil ではない**
>
> インターフェースが `nil` になるのは、**値と型の両方が空** のときだけです。`var p *MyError = nil` を `error` として返すと、受け取った側の `err != nil` が true になってしまいます。これは Go で非常によくあるバグです ([Go FAQ](https://go.dev/doc/faq#nil_error))。
>
> ```go
> func returnsError() error {
> 	var p *MyError = nil
> 	if bad() {
> 		p = ErrBad
> 	}
> 	return p // ❌ 常に non-nil の error になる。成功時は `return nil` と書く
> }
> ```

値も型も入っていない **nil インターフェース** に対してメソッドを呼ぶと、どのメソッドを呼ぶか決められないので、実行時エラー (panic) になります。

### 空インターフェース any

メソッドを 1 つも持たないインターフェース `interface{}` は、**どんな型の値でも保持できます**。Go 1.18 から、別名の `any` が使えるようになりました。

```go
var v any
v = 42
v = "hello"
v = []int{1, 2}
```

`fmt.Println(a ...any)` のように、どんな型の値でも受け取る必要がある場面で使います。ただし型による保護が失われるので、**できるだけ使わない** ようにしましょう。型を問わない処理を書きたいだけなら、ジェネリクス (第6章) のほうが適していることが多いです。

## 5.4 型アサーションと型 switch

インターフェース値から具体的な値を取り出すには、**型アサーション** を使います。

```go
package main

import "fmt"

func main() {
	var i any = "hello"

	s := i.(string) // 型が違うと panic する
	fmt.Println(s)

	s, ok := i.(string) // カンマ ok の形なら panic しない
	fmt.Println(s, ok)

	f, ok := i.(float64)
	fmt.Println(f, ok) // 失敗したときはゼロ値と false が返る

	// インターフェースへのアサーションもできる: 「String メソッドを持っているか」
	_, isStringer := i.(fmt.Stringer)
	fmt.Println(isStringer)
}
```

```text
hello
hello true
0 false
false
```

複数の型で処理を分けたいときは、**型 switch** を使います。

```go
package main

import (
	"fmt"
	"strconv"
)

type Celsius float64

func (c Celsius) String() string { return strconv.FormatFloat(float64(c), 'f', 1, 64) + "℃" }

func describe(v any) string {
	switch x := v.(type) { // x は各 case の型になる
	case nil:
		return "nil"
	case int, int64: // 型を複数並べると、x は any のまま
		return fmt.Sprintf("integer %v", x)
	case string:
		return "string of length " + strconv.Itoa(len(x))
	case fmt.Stringer: // インターフェースも case に書ける
		return "Stringer: " + x.String()
	case error:
		return "error: " + x.Error()
	default:
		return fmt.Sprintf("other %T", x)
	}
}

func main() {
	for _, v := range []any{nil, 42, "héllo", Celsius(36.5), fmt.Errorf("oops"), 3.14} {
		fmt.Println(describe(v))
	}
}
```

```text
nil
integer 42
string of length 6
Stringer: 36.5℃
error: oops
other float64
```

(`"héllo"` の長さが 6 になるのは、`é` が UTF-8 で 2 バイトだからです。)

### インターフェースを満たすかをコンパイル時に確かめる

型がインターフェースを満たしているかどうかは、通常は使う場所で自動的にチェックされます。しかし、たとえば `encoding/json` は実行時に `json.Marshaler` を実装しているか調べるので、実装を間違えても **コンパイルエラーにならず、ただ使われないだけ** です。こうした場合は、空白識別子を使って明示的にチェックします。

```go
var _ json.Marshaler = (*RawMessage)(nil) // *RawMessage が json.Marshaler を満たさなければコンパイルエラー
```

5.2 節の規則も、このチェックで確かめられます。

<!-- nocheck -->
```go
package main

type Scaler interface{ Scale(float64) }

type Vertex struct{ X, Y float64 }

func (v *Vertex) Scale(f float64) { v.X *= f; v.Y *= f }

var _ Scaler = (*Vertex)(nil) // ✅ *Vertex は Scaler を満たす
var _ Scaler = Vertex{}       // ❌ コンパイルエラー: Vertex does not implement Scaler (method Scale has pointer receiver)

func main() {}
```

## 5.5 標準ライブラリの重要なインターフェース

### fmt.Stringer

```go
type Stringer interface {
	String() string
}
```

`fmt` パッケージは、表示する値が `Stringer` を満たしていれば `String()` の戻り値を使います。

```go
package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s (%d歳)", p.Name, p.Age)
}

func main() {
	a := Person{"Arthur Dent", 42}
	z := Person{"Zaphod Beeblebrox", 9001}
	fmt.Println(a, z)
	fmt.Printf("%v | %s | %+v\n", a, a, []Person{a})
}
```

```text
Arthur Dent (42歳) Zaphod Beeblebrox (9001歳)
Arthur Dent (42歳) | Arthur Dent (42歳) | [Arthur Dent (42歳)]
```

> ⚠️ `String()` の中で `fmt.Sprintf("%v", p)` のように **自分自身を `%v` や `%s` で書式化すると、`String()` が再び呼ばれて無限再帰になります**。`Sprintf("%v", string(m))` のように、`String()` を持たない型に変換してから渡しましょう。

### error

```go
type error interface {
	Error() string
}
```

組み込みのインターフェースです。詳しくは第7章で扱います。

### io.Reader と io.Writer

Go で最も重要なインターフェースと言えるのが、メソッドを 1 つだけ持つこの 2 つです。

```go
type Reader interface {
	Read(p []byte) (n int, err error) // p に最大 len(p) バイト読み込み、読んだバイト数を返す。終端では io.EOF
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
```

ファイル、ネットワーク接続、HTTP のリクエストボディとレスポンス、`bytes.Buffer`、`strings.Reader`、gzip、暗号化ストリームなど、**数え切れないほどの型がこの 2 つを実装** しています。そのため、どの型も同じ道具で組み合わせられます。

```go
package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	r := strings.NewReader("Hello, Reader!")
	b := make([]byte, 8)
	for {
		n, err := r.Read(b)
		fmt.Printf("n = %v err = %v b[:n] = %q\n", n, err, b[:n])
		if err == io.EOF {
			break
		}
	}
}
```

```text
n = 8 err = <nil> b[:n] = "Hello, R"
n = 6 err = <nil> b[:n] = "eader!"
n = 0 err = EOF b[:n] = ""
```

Reader をラップして別の Reader を作るのは、Go の定番の設計です。

```go
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	// 書き込み: 文字列 → gzip 圧縮 → バッファ
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	original := strings.Repeat("gopher ", 100)
	io.Copy(zw, strings.NewReader(original))
	zw.Close()
	fmt.Println("compressed:", buf.Len() < len(original)) // 700 バイトが数十バイトになる

	// 読み込み: バッファ → gzip 展開 → 先頭 20 バイトだけ → 標準出力
	zr, _ := gzip.NewReader(&buf)
	io.Copy(os.Stdout, io.LimitReader(zr, 20))
	fmt.Println()
}
```

```text
compressed: true
gopher gopher gopher
```

### image.Image

```go
package image

type Image interface {
	ColorModel() color.Model
	Bounds() Rectangle
	At(x, y int) color.Color
}
```

このインターフェースを実装すれば、`png.Encode` などの標準のエンコーダにそのまま渡せます (演習 5-4)。

## 5.6 インターフェースの設計

### 小さく保つ

> *The bigger the interface, the weaker the abstraction.* — Go Proverbs

インターフェースは、メソッドが 1〜3 個程度の小さなものが理想です。メソッドが 1 つのインターフェースは、**メソッド名に -er を付けた名前** にするのが慣習です (`Reader`、`Writer`、`Formatter`、`Stringer`、`Closer`)。

### 汎用性: 具体型ではなくインターフェースを返す場合

ある型が **インターフェースを実装するためだけに存在し**、それ以外に公開するメソッドがないなら、コンストラクタは **インターフェースを返す** ようにして、型そのものは非公開にできます。こうすると、実装を差し替えても利用者のコードは変わりません。

```go
// 標準ライブラリの例: 戻り値は hash.Hash32 インターフェース
func NewIEEE() hash.Hash32  // hash/crc32
func New() hash.Hash32      // hash/adler32
```

一方で、一般的な指針は **「インターフェースを受け取り、具体型 (構造体) を返す」** です。具体型を返しておけば、利用者はその型の全メソッドを使えます。後からメソッドを追加しても互換性が壊れません。上のようにインターフェースを返すのは、実装を隠すこと自体に意味がある場合の例外です。

### メソッドを持てるのは構造体だけではない — http.HandlerFunc

`net/http` の `Handler` インターフェースは、どんな型でも実装できます。

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

Effective Go では、さまざまな型でこのインターフェースを実装する例を紹介しています。

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// 1. 構造体で実装する
type Counter struct{ n int }

func (c *Counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.n++
	fmt.Fprintf(w, "counter = %d", c.n)
}

// 2. int 型で実装する (状態がカウンタだけなら構造体は要らない)
type IntCounter int

func (c *IntCounter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*c++
	fmt.Fprintf(w, "int counter = %d", *c)
}

// 3. 関数型で実装する: 標準ライブラリの http.HandlerFunc の仕組み
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r) // 自分自身 (関数) を呼ぶだけ
}

func hello(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "hello") }

func main() {
	handlers := []http.Handler{
		&Counter{},
		new(IntCounter),
		HandlerFunc(hello), // 普通の関数を型変換するだけで Handler になる
	}
	for _, h := range handlers {
		for range 2 {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			fmt.Println(rec.Body.String())
		}
	}
}
```

```text
counter = 1
counter = 2
int counter = 1
int counter = 2
hello
hello
```

3 つ目のように **関数型にメソッドを定義して、関数をインターフェースに合わせる** 手法は、アダプタとして標準ライブラリやミドルウェアで多用されています (第15章)。

### 型変換でメソッドを切り替える — sort.Interface の例

同じデータでも、型を変換すると別のメソッドセットを使えます。次の例 (Effective Go の例を元にしています) では、`[]int` を元にした `Sequence` 型に `sort.Interface` を実装し、ソートしてから表示しています。

```go
package main

import (
	"fmt"
	"sort"
)

type Sequence []int

// sort.Interface の実装
func (s Sequence) Len() int           { return len(s) }
func (s Sequence) Less(i, j int) bool { return s[i] < s[j] }
func (s Sequence) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s Sequence) String() string {
	s = append(Sequence(nil), s...) // 呼び出し元のデータを変更しないようにコピーする
	sort.Sort(s)
	return fmt.Sprint([]int(s)) // []int に変換してから渡すので、String() は再び呼ばれない
}

func main() {
	seq := Sequence{5, 2, 8, 1}
	fmt.Println(seq)
	fmt.Println([]int(seq)) // 元のデータは変わっていない
}
```

```text
[1 2 5 8]
[5 2 8 1]
```

> 💡 Go 1.21 以降は、`slices.Sort(s)` や `slices.SortFunc(s, cmp)` を使うほうが簡単です。`sort.Interface` は、既存のコードで今でもよく見かけるので知っておきましょう。

## 5.7 埋め込み (Embedding)

Go には継承がありません。代わりに、ある型を別の型の中に **名前を付けずに埋め込む** ことで、メソッドを外側の型に **昇格 (promote)** できます。

### インターフェースの埋め込み

```go
type Reader interface{ Read(p []byte) (n int, err error) }
type Writer interface{ Write(p []byte) (n int, err error) }

// ReadWriter は Reader と Writer の両方のメソッドを持つ
type ReadWriter interface {
	Reader
	Writer
}
```

### 構造体の埋め込み

次の `Job` の例は Effective Go の例を元にしています。

```go
package main

import (
	"log"
	"os"
)

type Job struct {
	Command     string
	*log.Logger // 型だけを書く (フィールド名は型名の Logger になる)
}

func NewJob(command string) *Job {
	return &Job{command, log.New(os.Stdout, "[job] ", 0)}
}

func main() {
	job := NewJob("backup")
	job.Println("starting", job.Command) // log.Logger のメソッドを Job から直接呼べる
	job.Logger.SetPrefix("[job2] ")      // 埋め込んだフィールドには型名でアクセスできる
	job.Printf("done")
}
```

```text
[job] starting backup
[job2] done
```

- 埋め込みは **継承ではなく委譲** です。`job.Println(...)` は `job.Logger.Println(...)` の省略形にすぎません。呼ばれたメソッドのレシーバは `Job` ではなく `*log.Logger` です。
- 外側の型で同じ名前のメソッドを定義すると、**外側のメソッドが優先** されます。それでも埋め込んだ型のメソッドは `job.Logger.Println` のように明示すれば呼べます。オーバーライドとは違い、**埋め込まれた側から外側のメソッドを呼ぶことはできません**。
- 同じ深さに同じ名前が 2 つあると、その名前を使った時点でコンパイルエラーになります。

標準ライブラリの `bufio.ReadWriter` は、`*bufio.Reader` と `*bufio.Writer` を埋め込んで `io.ReadWriter` を満たしています。

```go
type ReadWriter struct {
	*Reader // *bufio.Reader
	*Writer // *bufio.Writer
}
```

### 一部のメソッドだけ差し替える

インターフェースを埋め込んだ構造体を作ると、**一部のメソッドだけを上書きし、残りはそのまま委譲する** ラッパーが簡単に作れます。

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// WriteHeader だけを上書きし、それ以外の Header() と Write() は埋め込んだ ResponseWriter に委譲する
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func main() {
	rec := &statusRecorder{ResponseWriter: httptest.NewRecorder(), status: 200}
	http.NotFound(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println("status:", rec.status)
}
```

```text
status: 404
```

第15章の Web API では、この手法でアクセスログを出すミドルウェアを作ります。

## ✍️ 演習

### 演習 5-1: Stringer

[`exercises/stringer`](../exercises/stringer/exercise.go): `type IPAddr [4]byte` に `String()` メソッドを実装し、`IPAddr{127, 0, 0, 1}` が `"127.0.0.1"` と表示されるようにしてください。

### 演習 5-2: Reader

[`exercises/reader`](../exercises/reader/exercise.go): `'A'` を無限に返し続ける `io.Reader` を実装してください。

### 演習 5-3: rot13Reader

[`exercises/rot13`](../exercises/rot13/exercise.go): 別の `io.Reader` をラップし、アルファベットを 13 文字ずらす (ROT13 暗号) `Rot13Reader` を実装してください。`"Lbh penpxrq gur pbqr!"` を読むと `"You cracked the code!"` になります。

ヒント: 内側の Reader の `Read` が返したバイト数 `n` までだけを変換します。

### 演習 5-4: Image

[`exercises/image`](../exercises/image/exercise.go): `image.Image` インターフェースを満たす `Image` 型を実装してください。

```bash
cd exercises
go test ./stringer/ ./reader/ ./rot13/ ./image/
```

解答例はそれぞれのディレクトリの `solution.go` にあります。
