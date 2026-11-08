# 第6章 ジェネリクスとイテレータ

Go 1.18 で **型パラメータ (ジェネリクス)** が、Go 1.23 で **range-over-func (イテレータ)** が導入されました。どちらも、型に依存しない処理を **型安全なまま** 書くための仕組みです。

## 6.1 型パラメータを持つ関数

関数名の後ろの `[ ]` に **型パラメータ** と、その型が満たすべき **制約** を書きます。

```go
package main

import "fmt"

// Index は s の中で x が最初に現れる位置を返す。見つからなければ -1
// comparable は == と != で比較できる型の集合を表す制約
func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		if v == x {
			return i
		}
	}
	return -1
}

func main() {
	si := []int{10, 20, 15, -10}
	fmt.Println(Index(si, 15)) // T=int と推論される

	ss := []string{"foo", "bar", "baz"}
	fmt.Println(Index(ss, "hello"))

	fmt.Println(Index[float64]([]float64{1.5}, 1.5)) // 型引数を明示することもできる
}
```

```text
2
-1
0
```

ジェネリクスがなかった頃は、型ごとに同じ関数をコピーして書くか、`any` で受け取って実行時に型アサーションするしかありませんでした。

## 6.2 制約

制約はインターフェースとして書きます。

| 制約 | 意味 |
| --- | --- |
| `any` | 制限なし (どの型でもよい) |
| `comparable` | `==` と `!=` で比較できる型。マップのキーに使える |
| `cmp.Ordered` | `<` や `>` で順序を比べられる型 (整数、浮動小数点数、文字列) |
| 自作のインターフェース | 型の集合 (union) やメソッドを指定できる |

```go
package main

import (
	"cmp"
	"fmt"
)

// 型の集合 (union) で制約を作る。~int は「元になる型が int の型」すべてを含む
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Sum[T Number](xs []T) T {
	var total T // T のゼロ値
	for _, x := range xs {
		total += x
	}
	return total
}

func MaxOf[T cmp.Ordered](first T, rest ...T) T {
	m := first
	for _, v := range rest {
		m = max(m, v) // max と min は Go 1.21 からの組み込み関数
	}
	return m
}

type Celsius float64 // 元になる型は float64 なので ~float64 に含まれる

func main() {
	fmt.Println(Sum([]int{1, 2, 3}))
	fmt.Println(Sum([]Celsius{36.5, 0.5}))
	fmt.Println(MaxOf(3, 9, 4), MaxOf("go", "rust", "c"))
	fmt.Println(cmp.Compare(1, 2), cmp.Compare("b", "a"))
}
```

```text
6
37
9 rust
-1 1
```

- `~` を付けないと、`Celsius` のように **自分で定義した型** は制約を満たしません。
- メソッドと型の集合を組み合わせた制約も書けます。たとえば `interface{ ~int; String() string }` は、「元になる型が int で、`String()` メソッドを持つ型」を表します。
- 型の集合を含むインターフェースは **制約としてだけ** 使えます。普通の変数の型には使えません。

## 6.3 ジェネリックな型

型の定義にも型パラメータを付けられます。

```go
package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // T のゼロ値は var で作る
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// 型パラメータを 2 つ持つ型
type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func main() {
	var s Stack[string] // ゼロ値のまま使える
	s.Push("a")
	s.Push("b")
	v, _ := s.Pop()
	fmt.Println(v, s.Len())

	p := Pair[string, int]{"age", 42}
	fmt.Printf("%+v\n", p)
}
```

```text
b 1
{Key:age Val:42}
```

### 制限事項

- **メソッドは、新しい型パラメータを持てません**。型に付けた型パラメータを使うことはできます。`func (s *Stack[T]) Map[U any](...)` のようなメソッドは書けないので、パッケージレベルの関数 `func Map[T, U any](s *Stack[T], ...)` として書きます。
- 型パラメータを持つ関数や型を使うときは、`Stack[string]` のように具体的な型を指定する (インスタンス化する) 必要があります。関数の型引数は、引数から推論できれば省略できます。

## 6.4 標準ライブラリのジェネリックなパッケージ

| パッケージ | 主な関数 |
| --- | --- |
| `slices` | `Sort`, `SortFunc`, `Index`, `Contains`, `Max`, `Min`, `Reverse`, `Compact`, `Insert`, `Delete`, `Equal`, `BinarySearch`, `Clone`, `Collect`, `Sorted` |
| `maps` | `Keys`, `Values`, `Clone`, `Copy`, `Equal`, `DeleteFunc`, `Collect` |
| `cmp` | `Compare`, `Less`, `Or`, `Ordered` |
| `sync` | `OnceValue[T]`, `OnceValues[T1, T2]` |
| `iter` | `Seq[V]`, `Seq2[K, V]`, `Pull` |

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type User struct {
	Name string
	Age  int
}

func main() {
	users := []User{{"carol", 35}, {"alice", 30}, {"bob", 30}}

	// 年齢の昇順、同じ年齢なら名前の昇順で並べる
	slices.SortFunc(users, func(a, b User) int {
		return cmp.Or(
			cmp.Compare(a.Age, b.Age),
			strings.Compare(a.Name, b.Name),
		)
	})
	fmt.Println(users)

	i, found := slices.BinarySearchFunc(users, 35, func(u User, age int) int {
		return cmp.Compare(u.Age, age)
	})
	fmt.Println(i, found)
}
```

```text
[{alice 30} {bob 30} {carol 35}]
2 true
```

## 6.5 いつジェネリクスを使うか

Go チームの Ian Lance Taylor による指針 ([When To Use Generics](https://go.dev/blog/when-generics)) をまとめると、次のとおりです。

**✅ 使うとよい場面**

- スライス・マップ・チャネルを扱う、**要素の型に依存しない** 関数 (`slices`、`maps` パッケージがまさにこれです)
- 汎用的なデータ構造 (連結リスト、木、ヒープ、キャッシュ、集合など)
- 処理の中身は同じで、型だけが違う関数を何度も書いていると気付いたとき

**❌ 使わないほうがよい場面**

- **引数のメソッドを呼ぶだけなら、普通のインターフェースで十分** です。`func Write[W io.Writer](w W)` ではなく `func Write(w io.Writer)` と書きます。
- 型ごとに **処理の中身が違う** 場合 (それはジェネリクスではなく、インターフェースのメソッドで実装すべきです)
- 「将来役に立つかもしれない」という理由での先回りの汎用化

> まずは普通の関数を書きましょう。同じコードを型だけ変えて 2 回、3 回と書いたときが、ジェネリクスに書き換えるタイミングです。

## 6.6 イテレータ (Go 1.23+)

`func(yield func(V) bool)` という形の関数を、**`for range` で回せる** ようになりました。

```go
package iter

type Seq[V any] func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)
```

```go
package main

import (
	"fmt"
	"iter"
)

// Countdown は n から 1 までを順に返すイテレータ
func Countdown(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := n; i >= 1; i-- {
			if !yield(i) { // ループ側で break されると false が返るので、すぐに終了する
				return
			}
		}
	}
}

// Enumerate はインデックス付きで値を返す (Seq2)
func Enumerate[T any](s []T) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
	}
}

func main() {
	for v := range Countdown(5) {
		if v == 2 {
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()

	for i, s := range Enumerate([]string{"a", "b"}) {
		fmt.Println(i, s)
	}
}
```

```text
5 4 3 
0 a
1 b
```

`for range` で回すと、ループの本体が `yield` 関数として渡されます。ループの中で `break` すると `yield` が `false` を返すので、**イテレータ側は必ず戻り値を確認して終了** しなければなりません。確認しないと実行時に panic します。

### 標準ライブラリとの組み合わせ

イテレータの最大の利点は、**標準ライブラリの関数を部品のように組み合わせられる** ことです。

```go
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	scores := map[string]int{"carol": 72, "alice": 90, "bob": 85}

	// マップのキーを取り出して → ソートしてスライスにする
	names := slices.Sorted(maps.Keys(scores))
	fmt.Println(names)

	// スライスを (index, value) のイテレータにして → マップに集める
	m := maps.Collect(slices.All([]string{"zero", "one"}))
	fmt.Println(m)

	// 文字列を分割した結果をイテレータで受け取る (Go 1.24+。スライスを作らない)
	for part := range strings.SplitSeq("a,b,c", ",") {
		fmt.Print(part, ";")
	}
	fmt.Println()
}
```

```text
[alice bob carol]
map[0:zero 1:one]
a;b;c;
```

### iter.Pull: 値を 1 つずつ取り出す

2 つのイテレータを並べて比較するときなど、`for range` で回すのではなく **必要なときに 1 つずつ値を取り出したい** 場合は、`iter.Pull` を使います。

```go
package main

import (
	"fmt"
	"iter"
	"slices"
)

func Equal[V comparable](a, b iter.Seq[V]) bool {
	nextA, stopA := iter.Pull(a)
	defer stopA()
	nextB, stopB := iter.Pull(b)
	defer stopB()
	for {
		va, okA := nextA()
		vb, okB := nextB()
		if okA != okB || va != vb {
			return false
		}
		if !okA {
			return true
		}
	}
}

func main() {
	fmt.Println(Equal(slices.Values([]int{1, 2}), slices.Values([]int{1, 2})))
	fmt.Println(Equal(slices.Values([]int{1, 2}), slices.Values([]int{1, 3})))
}
```

```text
true
false
```

## ✍️ 演習: ジェネリックな関数と型

[`exercises/generics`](../exercises/generics/exercise.go) を実装してください。

1. `Map[T, U any](xs []T, f func(T) U) []U`
2. `Filter[T any](xs []T, keep func(T) bool) []T`
3. `Reduce[T, A any](xs []T, init A, f func(A, T) A) A`
4. ゼロ値のまま使える単方向連結リスト `List[T]`。`Push(v T)` と、イテレータを返す `All() iter.Seq[T]` を持たせます。

```bash
cd exercises
go test ./generics/
```

第17章のハンズオンでは、ジェネリクスとイテレータを使って、実用的な LRU キャッシュライブラリを作ります。
