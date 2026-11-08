# 第3章 インターフェースとジェネリクス

## 3.1 インターフェースは「暗黙的に」実装する

```go
type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }
func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }
func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

func total(shapes ...Shape) float64 {
	var sum float64
	for _, s := range shapes {
		sum += s.Area()
	}
	return sum
}
```

`Rect` は `implements Shape` と宣言していません。**メソッドが揃っていれば自動的に満たす** (構造的部分型) のが Go の特徴です。これにより、

- 実装側がインターフェースを知らなくてよい → 依存の向きを自由に決められる
- 既存の型 (標準ライブラリや外部パッケージの型) にも後からインターフェースを当てはめられる

### 標準ライブラリの小さなインターフェース

Go で最も重要なインターフェースは、メソッド 1 つだけの `io.Reader` / `io.Writer` です。

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

ファイル、ネットワーク接続、HTTP ボディ、圧縮ストリーム、`bytes.Buffer`、`strings.Reader`… すべてがこれを満たすため、組み合わせ自在です。

```go
// ファイルを gzip 圧縮しながら標準出力へ
f, _ := os.Open("access.log")
zw := gzip.NewWriter(os.Stdout)
io.Copy(zw, f)
zw.Close()
```

### 設計指針

| 指針 | 説明 |
| --- | --- |
| **小さく保つ** | *The bigger the interface, the weaker the abstraction.* |
| **使う側で定義する** | インターフェースは実装パッケージではなく利用パッケージに置く ([Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#interfaces)) |
| **Accept interfaces, return structs** | 引数はインターフェースで受け、戻り値は具体型で返す |
| **必要になるまで作らない** | 実装が 1 つしかないのに抽象化しない。テストで差し替えたくなったときが作り時 |

コンパイル時に実装を保証したいときのイディオム:

```go
var _ Shape = Rect{}            // Rect が Shape を満たさないとコンパイルエラー
var _ io.Writer = (*MyWriter)(nil)
```

> 📦 **OSS ではこう書かれている**: GitHub CLI (`cli/cli`) は、標準入出力を `IOStreams` 構造体 (`In io.Reader`, `Out io.Writer`, `ErrOut io.Writer`) にまとめて各コマンドに注入し、テストでは `bytes.Buffer` に差し替えています。本書の第10章 TODO CLI でも同じ設計を採用しています。

## 3.2 型アサーションと型スイッチ

```go
var v any = "hello" // any は interface{} の別名

s, ok := v.(string) // 型アサーション (comma ok)

switch x := v.(type) {
case string:
	fmt.Println("string", len(x))
case int:
	fmt.Println("int", x)
case fmt.Stringer:
	fmt.Println(x.String())
default:
	fmt.Println("unknown")
}
```

## 3.3 ジェネリクス (Go 1.18+)

型パラメータで、型に依存しない関数やデータ構造を書けます。

```go
// 制約 cmp.Ordered は < で比較できる型の集合
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

Max(3, 7)         // 型推論で T=int
Max("go", "rust") // T=string
```

### 型制約を自作する

```go
type Number interface {
	~int | ~int64 | ~float64 // ~ は「基底型が int の独自型」も含む
}

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}
```

### ジェネリックな型

```go
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}
```

### いつジェネリクスを使うか

Go チームの指針 ([When To Use Generics](https://go.dev/blog/when-generics)) の要約:

- ✅ スライス・マップ・チャネルに対する汎用処理 (`slices`, `maps` パッケージがまさにこれ)
- ✅ 汎用データ構造 (連結リスト、木、キャッシュなど)
- ❌ メソッドを呼ぶだけならインターフェースで十分
- ❌ 型ごとに実装が違うならジェネリクスではない

→ 第13章のハンズオンで、ジェネリクスを使った LRU キャッシュを実装します。

## 3.4 イテレータ (Go 1.23+ range-over-func)

関数を `for range` で回せるようになりました。

```go
import "iter"

// 0 から n-1 までの偶数を返すイテレータ
func Evens(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i += 2 {
			if !yield(i) { // false が返ったら (break されたら) 止める
				return
			}
		}
	}
}

for v := range Evens(10) {
	fmt.Println(v) // 0 2 4 6 8
}
```

標準ライブラリでも `maps.Keys`, `slices.All`, `strings.SplitSeq` (1.24) など、イテレータを返す関数が増えています。

## ✍️ 練習問題

1. `Filter[T any](xs []T, keep func(T) bool) []T` を実装しよう。
2. 前章の `Rectangle` / `Circle` を `Shape` インターフェースで扱い、面積の合計を出そう。
3. `iter.Seq[int]` を返すフィボナッチ数列イテレータを書き、`for` で最初の 10 個を表示しよう。
