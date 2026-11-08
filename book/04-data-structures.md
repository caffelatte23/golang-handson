# 第4章 データ構造 — ポインタ・構造体・配列・スライス・マップ・関数値

## 4.1 ポインタ

ポインタは値のメモリアドレスを保持します。`*T` は「T 型の値を指すポインタ」の型で、ゼロ値は `nil` です。

```go
package main

import "fmt"

func main() {
	i, j := 42, 2701

	p := &i         // & で i のアドレスを取得する
	fmt.Println(*p) // * でポインタが指す先の値を読む (デリファレンス)
	*p = 21         // ポインタを通して i を書き換える
	fmt.Println(i)

	p = &j
	*p = *p / 37
	fmt.Println(j)
}
```

```text
42
21
73
```

- C と違い、**ポインタ演算 (`p++` など) はできません**。そのためメモリ安全性が保たれます。
- 関数の **引数は常に値渡し (コピー)** です。関数の中で呼び出し元の変数を変更したいときは、ポインタを渡します。

```go
package main

import "fmt"

func incrementValue(n int)    { n++ }
func incrementPointer(n *int) { *n++ }

func main() {
	x := 1
	incrementValue(x)
	fmt.Println(x) // 1: コピーが変更されただけ
	incrementPointer(&x)
	fmt.Println(x) // 2
}
```

```text
1
2
```

- ローカル変数のアドレスを関数から返してもかまいません。必要ならコンパイラがその変数をヒープに置きます (エスケープ解析)。C のようにスタックの領域が解放されてダングリングポインタになることはありません。

## 4.2 構造体

構造体 (`struct`) はフィールドの集まりです。

```go
package main

import "fmt"

type Vertex struct {
	X, Y int
}

func main() {
	v := Vertex{1, 2}
	v.X = 4 // フィールドにはドットでアクセスする
	fmt.Println(v)

	p := &v
	p.X = 1e9 // (*p).X と書く必要はない。ポインタからも直接アクセスできる
	fmt.Println(v)

	// 構造体リテラル
	v1 := Vertex{X: 1}  // フィールド名を指定する。省略した Y はゼロ値
	v2 := Vertex{}      // X:0, Y:0
	v3 := &Vertex{1, 2} // *Vertex 型
	fmt.Println(v1, v2, *v3)
	fmt.Printf("%v %+v %#v\n", v1, v1, v1)

	// 構造体はすべてのフィールドが比較可能なら == で比較できる
	fmt.Println(v1 == Vertex{X: 1, Y: 0})
}
```

```text
{4 2}
{1000000000 2}
{1 0} {0 0} {1 2}
{1 0} {X:1 Y:0} main.Vertex{X:1, Y:0}
true
```

- `Vertex{1, 2}` のようにフィールド名を省いて順番どおりに書く形は、**フィールドが増えると壊れます**。他のパッケージの構造体には必ず `Vertex{X: 1, Y: 2}` の形を使いましょう (`go vet` の composites チェックで警告されます)。

### 構造体タグ

フィールドの後ろに書く文字列を **タグ** と呼びます。`encoding/json` などのパッケージがリフレクションで読み取り、変換方法の指示として使います。

```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"` // 空なら出力しない
	Password string `json:"-"`               // 出力しない
	note     string // 非公開フィールドは JSON に含まれない
}

func main() {
	b, _ := json.Marshal(User{ID: 1, Name: "Gopher", Password: "secret", note: "x"})
	fmt.Println(string(b))

	var u User
	_ = json.Unmarshal([]byte(`{"id":2,"name":"Ken","email":"ken@example.com"}`), &u)
	fmt.Printf("%+v\n", u)
}
```

```text
{"id":1,"name":"Gopher"}
{ID:2 Name:Ken Email:ken@example.com Password: note:}
```

### 無名構造体

その場限りで使う型は、名前を付けずに定義できます。テーブル駆動テストや JSON の一時的な受け皿によく使います。

```go
point := struct{ X, Y int }{10, 20}
```

## 4.3 配列

`[n]T` は T 型の値を n 個持つ配列です。**長さも型の一部** なので、`[3]int` と `[4]int` は別の型です。

```go
package main

import "fmt"

func main() {
	var a [2]string
	a[0], a[1] = "Hello", "World"
	fmt.Println(a, len(a))

	primes := [6]int{2, 3, 5, 7, 11, 13}
	b := primes // 配列は「値」なので、代入すると全体がコピーされる
	b[0] = 999
	fmt.Println(primes[0], b[0])

	auto := [...]string{"a", "b", "c"} // 長さを要素数から推論させる
	fmt.Println(len(auto))
}
```

```text
[Hello World] 2
2 999
3
```

配列は値なので、関数に渡すとコピーされます。大きな配列を渡すとコストがかかります。長さを変えることもできません。そのため Go のコードで配列を直接使うことはまれで、ほとんどの場面で **スライス** を使います。

## 4.4 スライス

`[]T` は T 型のスライスです。スライスは **配列の一部分を指す軽量な参照** で、内部は 3 つの値でできています。

```text
スライス = { 配列へのポインタ, 長さ len, 容量 cap }
```

```go
package main

import "fmt"

func main() {
	primes := [6]int{2, 3, 5, 7, 11, 13}

	var s []int = primes[1:4] // 添字 1 から 3 まで (上限の 4 は含まない)
	fmt.Println(s)

	// スライスはデータを持たない。元の配列を共有しているだけ
	s[0] = 999
	fmt.Println(primes)

	// 上限・下限は省略できる
	fmt.Println(primes[:2], primes[4:], len(primes[:]))
}
```

```text
[3 5 7]
[2 999 5 7 11 13]
[2 999] [11 13] 6
```

### スライスリテラル

```go
q := []int{2, 3, 5, 7}             // 配列を作り、それを指すスライスを返す
names := []string{"Go", "Rust"}
points := []struct{ X, Y int }{{1, 2}, {3, 4}} // 要素の型名は省略できる
```

### 長さ (len) と容量 (cap)

- **長さ**: スライスに含まれる要素の数です。
- **容量**: スライスの先頭から、元になる配列の末尾までの要素の数です。

```go
package main

import "fmt"

func printSlice(s []int) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func main() {
	s := []int{2, 3, 5, 7, 11, 13}
	printSlice(s)

	s = s[:0] // 長さを 0 に
	printSlice(s)

	s = s[:4] // 容量の範囲内なら長さを伸ばせる
	printSlice(s)

	s = s[2:] // 先頭の 2 要素を落とす (容量も減る)
	printSlice(s)
}
```

```text
len=6 cap=6 [2 3 5 7 11 13]
len=0 cap=6 []
len=4 cap=6 [2 3 5 7]
len=2 cap=4 [5 7]
```

### nil スライス

スライスのゼロ値は `nil` で、長さも容量も 0 です。**nil スライスにもそのまま `append` や `len` や `range` が使える** ので、空のスライスとして扱えます。

```go
var s []int
fmt.Println(s, len(s), cap(s), s == nil) // [] 0 0 true
s = append(s, 1)                         // nil にも append できる
```

### make でスライスを作る

```go
a := make([]int, 5)    // len=5 cap=5 [0 0 0 0 0]
b := make([]int, 0, 5) // len=0 cap=5 []
```

最終的な要素数が分かっているときは、`make([]T, 0, n)` で容量を確保しておくと、`append` のたびに配列を作り直す処理 (再割り当て) が起きません。

### append

```go
package main

import "fmt"

func main() {
	var s []int
	for i := range 10 {
		s = append(s, i) // 戻り値を必ず代入する
	}
	fmt.Println(s, len(s), cap(s) >= len(s))

	x := []int{1, 2, 3}
	y := []int{4, 5, 6}
	x = append(x, y...) // ... でスライスを可変長引数として展開する
	fmt.Println(x)
}
```

```text
[0 1 2 3 4 5 6 7 8 9] 10 true
[1 2 3 4 5 6]
```

`append` は、容量が足りなくなると **より大きな配列を新しく確保して要素をコピー** し、その配列を指すスライスを返します。元のスライス変数は古い配列を指したままなので、**`s = append(s, ...)` のように戻り値を必ず代入** してください。容量がどのように増えるか (2 倍、1.25 倍など) はランタイムの実装によって決まり、Go のバージョンによっても変わります。ループの中で `cap(s)` を表示して確かめてみましょう。

> ⚠️ **配列の共有によるバグ**
>
> ```go
> a := []int{1, 2, 3, 4}
> b := a[:2]         // b は a と同じ配列を共有している (cap=4)
> b = append(b, 99)  // 容量に余裕があるので、a[2] が書き換わる!
> fmt.Println(a)     // [1 2 99 4]
> ```
>
> スライスを他の場所に渡して独立に変更させたいときは、`slices.Clone(a)` でコピーするか、`a[:2:2]` (フルスライス式。3 つ目の数で容量を制限する) を使います。

### copy と slices パッケージ

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	src := []int{3, 1, 2}
	dst := make([]int, len(src))
	n := copy(dst, src) // コピーした要素数を返す
	fmt.Println(n, dst)

	s := []int{5, 2, 8, 1}
	slices.Sort(s)
	fmt.Println(s, slices.Contains(s, 8), slices.Index(s, 5))
	slices.Reverse(s) // その場で逆順にする
	fmt.Println(s, slices.Max(s), slices.Equal(s, []int{8, 5, 2, 1}))
}
```

```text
3 [3 1 2]
[1 2 5 8] true 2
[8 5 2 1] 8 true
```

`slices` パッケージ (Go 1.21+) には `Sort`、`SortFunc`、`BinarySearch`、`Index`、`Contains`、`Insert`、`Delete`、`Compact`、`Clone`、`Equal`、`Max`、`Min` など、スライス操作の定番関数がそろっています。自分で書く前に確認しましょう。

### 2 次元スライス

```go
package main

import "fmt"

func main() {
	// 方法1: 各行を個別に make する (行ごとに長さを変えられる)
	board := make([][]string, 3)
	for i := range board {
		board[i] = make([]string, 3)
		for j := range board[i] {
			board[i][j] = "_"
		}
	}
	board[0][0], board[1][1], board[2][2] = "X", "O", "X"
	for _, row := range board {
		fmt.Println(row)
	}

	// 方法2: 1 回の割り当てで全体を確保し、行ごとに切り出す (割り当て回数が少ない)
	const w, h = 4, 2
	pixels := make([]uint8, w*h)
	picture := make([][]uint8, h)
	for i := range picture {
		picture[i], pixels = pixels[:w], pixels[w:]
	}
	fmt.Println(len(picture), len(picture[0]))
}
```

```text
[X _ _]
[_ O _]
[_ _ X]
2 4
```

## 4.5 マップ

`map[K]V` はキーと値を対応付けるハッシュテーブルです。キーには **`==` で比較できる型** (数値、文字列、ポインタ、配列、比較可能な構造体など) を使えます。スライス、マップ、関数はキーにできません。また、キーの型をインターフェース (`any` など) にした場合、比較できない値 (スライスなど) をキーとして入れると **実行時に panic** します。

```go
package main

import "fmt"

type Vertex struct{ Lat, Long float64 }

func main() {
	m := map[string]Vertex{
		"Bell Labs": {40.68433, -74.39967}, // 値の型名は省略できる
		"Google":    {37.42202, -122.08408},
	}
	fmt.Println(m["Bell Labs"])

	ages := make(map[string]int)
	ages["alice"] = 31 // 追加・更新
	ages["bob"]++      // 存在しないキーはゼロ値として読めるので ++ も書ける
	delete(ages, "alice")

	v, ok := ages["carol"] // カンマ ok イディオム: 存在するかを確かめる
	fmt.Println(v, ok, len(ages))

	if age, ok := ages["bob"]; ok {
		fmt.Println("bob is", age)
	}
}
```

```text
{40.68433 -74.39967}
0 false 1
bob is 1
```

- **存在しないキーを読むとゼロ値が返ります**。値が「0 だった」のか「キーが無かった」のかを区別したいときは、カンマ ok を使います。
- **nil マップに書き込むと panic します**。`var m map[string]int` の後には `make` が必要です。読み取りは nil マップでもできます。
- **`range` で取り出す順番はランダムです**。順番が必要なときはキーをソートします。
- マップは参照型のように振る舞います。関数に渡して中身を変更すると、呼び出し元から見ても変わります。
- 複数の goroutine から同時に書き込むと、ランタイムがエラーを出してプログラムが停止します (第8章)。

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	stock := map[string]int{"banana": 3, "apple": 5, "cherry": 1}

	// キーを並べ替えてから回す (Go 1.23+ のイテレータを使う書き方。イテレータは第6章で説明する)
	for _, k := range slices.Sorted(maps.Keys(stock)) {
		fmt.Println(k, stock[k])
	}

	// 集合 (set) は map[T]bool か map[T]struct{} で表す
	seen := map[string]bool{}
	for _, w := range []string{"go", "is", "go"} {
		seen[w] = true
	}
	fmt.Println(len(seen), seen["go"], seen["rust"])
}
```

```text
apple 5
banana 3
cherry 1
2 true false
```

## 4.6 new と make

Go にはメモリを確保する組み込み関数が 2 つあります。

| | `new(T)` | `make(T, args)` |
| --- | --- | --- |
| 使える型 | 任意の型 | **スライス・マップ・チャネルだけ** |
| 戻り値 | `*T` (ゼロ値で初期化された領域へのポインタ) | `T` (**初期化済みで、すぐ使える状態** の値) |

```go
p := new([]int)       // *[]int。指している先は nil スライスなので、ほとんど役に立たない
v := make([]int, 100) // []int。長さ 100 の、すぐ使えるスライス
```

スライス・マップ・チャネルは、内部のデータ構造を初期化しないと使えません。そのための関数が `make` です。実際のコードでは `new(T)` の代わりに `&T{}` と書くことが多いです。

### 複合リテラルとコンストラクタ

```go
func NewFile(fd int, name string) *File {
	if fd < 0 {
		return nil
	}
	return &File{fd: fd, name: name} // ローカル変数のアドレスを返してもよい
}
```

型名の後に `{...}` と書いて値を作る式を **複合リテラル** と呼びます。`&` を付けるとポインタが得られます。Go にはコンストラクタの構文がないので、`NewXxx` という名前の関数を作るのが慣習です。パッケージの主要な型が 1 つだけなら、`ring.New` のように単に `New` とします。

## 4.7 関数値とクロージャ

関数も値です。変数に入れたり、引数や戻り値として受け渡したりできます。

```go
package main

import (
	"fmt"
	"math"
	"strings"
)

func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

func main() {
	hypot := func(x, y float64) float64 { // 関数リテラル (無名関数)
		return math.Sqrt(x*x + y*y)
	}
	fmt.Println(hypot(5, 12))
	fmt.Println(compute(hypot), compute(math.Pow))

	// 標準ライブラリでも関数値を受け取る API が多い
	fmt.Println(strings.Map(func(r rune) rune {
		if r == 'o' {
			return '0'
		}
		return r
	}, "gopher go"))
}
```

```text
13
5 81
g0pher g0
```

### クロージャ

関数リテラルは、その外側で宣言された変数を **捕捉 (キャプチャ)** して使えます。捕捉した変数は関数呼び出しが終わった後も残り、その関数と結び付いたままになります。

```go
package main

import "fmt"

func adder() func(int) int {
	sum := 0 // 返された関数ごとに別の sum を持つ
	return func(x int) int {
		sum += x
		return sum
	}
}

func main() {
	pos, neg := adder(), adder()
	for i := range 5 {
		fmt.Println(pos(i), neg(-2*i))
	}
}
```

```text
0 0
1 -2
3 -6
6 -12
10 -20
```

## 4.8 出力 (fmt)

`fmt` パッケージの関数は、出力先と書式によって次のように使い分けます。

| 関数 | 出力先 |
| --- | --- |
| `Print` / `Println` / `Printf` | 標準出力 |
| `Sprint` / `Sprintln` / `Sprintf` | 文字列として返す |
| `Fprint` / `Fprintln` / `Fprintf` | 任意の `io.Writer` (ファイル、HTTP レスポンス、`os.Stderr` など) |
| `Errorf` | `error` を作る (第7章) |

```go
package main

import (
	"fmt"
	"os"
)

type T struct {
	A int
	B string
	C []float64
}

func main() {
	t := &T{7, "go", []float64{1.5, 2}}
	fmt.Printf("%v\n", t)  // デフォルトの形式
	fmt.Printf("%+v\n", t) // フィールド名付き
	fmt.Printf("%#v\n", t) // Go の構文として表示
	fmt.Printf("%T\n", t)  // 型

	fmt.Printf("%d %5d %-5d| %05d %x %o %b\n", 42, 42, 42, 42, 255, 8, 5)
	fmt.Printf("%f %.2f %8.3f %e %g\n", 3.14159, 3.14159, 3.14159, 1234.5678, 0.000012)
	fmt.Printf("%s %q %v %c %U\n", "go", "go", true, 'G', 'G')

	fmt.Fprintln(os.Stdout, "to stdout via Fprintln")
	s := fmt.Sprintf("%03d-%s", 7, "abc")
	fmt.Println(s)
	fmt.Println(fmt.Sprint("a", 1, 2, "b"), fmt.Sprintln("a", 1, 2, "b"))
}
```

```text
&{7 go [1.5 2]}
&{A:7 B:go C:[1.5 2]}
&main.T{A:7, B:"go", C:[]float64{1.5, 2}}
*main.T
42    42 42   | 00042 ff 10 101
3.141590 3.14    3.142 1.234568e+03 1.2e-05
go "go" true G U+0047
to stdout via Fprintln
007-abc
a1 2b a 1 2 b
```

- `%v` は、どんな値でもそれらしい形で表示できる万能の書式です。マップは **キーでソートして** 表示されるので、テストの出力が安定します。
- `Sprint` は、隣り合う引数が **どちらも文字列でないときだけ** 間に空白を入れます。`Sprintln` は常に空白を入れ、最後に改行を付けます。
- 型に `String() string` メソッドを定義すると、`%v` や `Println` での表示を自分で決められます (第5章)。
- 書式指定と引数の数や型が合っていないと、`go vet` が警告してくれます。

## ✍️ 演習

### 演習 4-1: 2 次元スライス (Pic)

[`exercises/pic`](../exercises/pic/exercise.go) の `Pic(dx, dy int) [][]uint8` を実装してください。長さ `dy` のスライスを返し、その各要素は長さ `dx` の `[]uint8` にします。`p[y][x]` の値は `x^y` (XOR) です。

```bash
cd exercises
go test ./pic/
go run ./pic/cmd   # 実装できたら pic.png が生成されるので、画像として見てみよう
```

`x^y` の代わりに `(x+y)/2` や `x*y` にすると、どんな模様になるでしょうか。

### 演習 4-2: 単語の出現回数 (WordCount)

[`exercises/wordcount`](../exercises/wordcount/exercise.go) の `WordCount(s string) map[string]int` を実装し、各単語の出現回数を数えてください。単語の分割には `strings.Fields` を使います。

```bash
go test ./wordcount/
```

### 演習 4-3: フィボナッチのクロージャ

[`exercises/fibonacci`](../exercises/fibonacci/exercise.go) の `Fibonacci() func() int` を実装してください。返される関数は、呼ぶたびにフィボナッチ数列の次の値 (0, 1, 1, 2, 3, 5, ...) を返します。

```bash
go test ./fibonacci/
```

解答例はそれぞれのディレクトリの `solution.go` にあります。
