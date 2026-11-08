# 第3章 制御構文 — for・if・switch・defer

Go の制御構文は少数です。ループは `for` だけ、条件分岐は `if` と `switch` だけです。そのぶん、それぞれの使い方が柔軟になっています。

## 3.1 for

Go のループは `for` だけです。条件式を括弧 `( )` で囲まないことと、本体の `{ }` を省略できないことが C や Java との違いです。

```go
package main

import "fmt"

func main() {
	// 1. C と同じ 3 つの部分からなる形
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)

	// 2. 初期化と後処理を省くと、他の言語の while と同じ
	n := 1
	for n < 100 {
		n *= 2
	}
	fmt.Println(n)

	// 3. 整数に対する range (Go 1.22+): 0, 1, 2
	for i := range 3 {
		fmt.Print(i, " ")
	}
	fmt.Println()
}
```

```text
45
128
0 1 2 
```

条件を書かずに `for { ... }` とすると無限ループになります。抜けるには `break` か `return` を使います。

### range

スライス・配列・文字列・マップ・チャネル・整数・イテレータ関数を順に取り出すには `for range` を使います。

```go
package main

import "fmt"

func main() {
	langs := []string{"Go", "Rust", "Zig"}

	for i, v := range langs { // インデックスと値
		fmt.Println(i, v)
	}
	for i := range langs { // インデックスだけ
		_ = i
	}
	for _, v := range langs { // 値だけ (_ は空白識別子で、受け取った値を捨てる)
		_ = v
	}
}
```

```text
0 Go
1 Rust
2 Zig
```

> ⚠️ Go 1.22 から、**ループ変数はイテレーションごとに新しく作られる** ようになりました。それ以前は全イテレーションで 1 つの変数を共有していたため、goroutine やクロージャから参照するときに `v := v` とコピーする必要がありました。古い記事でこの書き方を見かけても、今は不要です。

### ラベル付き break / continue

入れ子のループで外側のループを抜けたいときは、ラベルを付けます。

```go
package main

import "fmt"

func main() {
	grid := [][]int{{1, 2, 3}, {4, -1, 6}, {7, 8, 9}}
outer:
	for y, row := range grid {
		for x, v := range row {
			if v < 0 {
				fmt.Printf("negative at (%d,%d)\n", x, y)
				break outer
			}
		}
	}
}
```

```text
negative at (1,1)
```

## 3.2 if

`for` と同じく条件を括弧で囲まず、`{ }` は必ず書きます。

```go
package main

import (
	"fmt"
	"math"
)

func pow(x, n, limit float64) float64 {
	// if の条件の前に短い文を書ける。v のスコープは if/else の中だけ
	if v := math.Pow(x, n); v < limit {
		return v
	} else {
		fmt.Printf("%g >= %g\n", v, limit)
	}
	// ここでは v は使えない
	return limit
}

func main() {
	fmt.Println(pow(3, 2, 10), pow(3, 3, 20))
}
```

```text
27 >= 20
9 20
```

`27 >= 20` が先に表示されるのは、`Println` の引数 (`pow` の呼び出し 2 つ) がすべて評価されてから表示されるからです。

### else を書かずに早めに return する

Go では「異常な場合を先に処理して return し、正常な処理を左端に沿って下へ書く」スタイルが好まれます。

```go
// ❌ ネストが深い
f, err := os.Open(name)
if err == nil {
	d, err := f.Stat()
	if err == nil {
		codeUsing(f, d)
	}
}

// ✅ エラーを先に処理して抜ける。正常な処理はインデントが浅いまま続く
f, err := os.Open(name)
if err != nil {
	return err
}
d, err := f.Stat()
if err != nil {
	f.Close()
	return err
}
codeUsing(f, d)
```

if ブロックが `return`・`break`・`continue` で終わるなら、`else` は書きません。

## 3.3 switch

```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("macOS")
	case "linux", "freebsd": // 1 つの case に複数の値を並べられる
		fmt.Println("Unix 系")
	default:
		fmt.Println(os)
	}
}
```

```text
Unix 系
```

(Linux や Playground で実行した場合の出力です。macOS では `macOS` と表示されます。)

C や Java との違い:

- **各 case の最後に break を書く必要がありません**。一致した case だけが実行されます。次の case まで続けて実行したいときだけ `fallthrough` を書きます。
- case に書く値は定数でなくてもかまいません。整数以外の値や式も書けます。
- case は **上から順に評価され、一致した時点で止まります**。それより下の case の式は評価されません。

### 条件のない switch

`switch {` と書くと `switch true {` と同じ意味になります。長い if-else-if の連鎖をすっきり書けます。

```go
package main

import "fmt"

func grade(score int) string {
	switch {
	case score >= 80:
		return "A"
	case score >= 60:
		return "B"
	default:
		return "C"
	}
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func main() {
	fmt.Println(grade(85), grade(70), grade(30))
	fmt.Println(unhex('7'), unhex('b'), unhex('F'))
}
```

```text
A B C
7 11 15
```

### switch の中の break

switch の中で `break` を書くと、抜けるのは **switch だけ** です。外側の for を抜けたいときはラベルを使います。

```go
Loop:
	for n := 0; n < len(src); n += size {
		switch {
		case src[n] < sizeOne:
			break Loop // for ループを抜ける
		case src[n] < sizeTwo:
			// ...
		}
	}
```

型によって処理を分ける **型 switch** は第5章で扱います。

## 3.4 defer

`defer` を付けた関数呼び出しは、**それを書いた関数が return する直前** に実行されます。

```go
package main

import "fmt"

func main() {
	defer fmt.Println("world")
	fmt.Println("hello")
}
```

```text
hello
world
```

### 典型的な使い方: 後片付け

```go
func Contents(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close() // 開けた直後に「閉じる」処理を予約する

	var result []byte
	buf := make([]byte, 100)
	for {
		n, err := f.Read(buf[0:])
		result = append(result, buf[0:n]...)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err // ここで return しても f は閉じられる
		}
	}
	return string(result), nil
}
```

ファイルを開く処理と閉じる処理を隣に書けるので、**閉じ忘れを防げます**。途中に return がいくつあっても確実に実行されます。`sync.Mutex` の `Unlock` にも同じように使います。

### defer の 3 つの規則

```go
package main

import "fmt"

func main() {
	// 規則1: 複数の defer は「後に登録したものから先に」実行される (LIFO)
	for i := range 3 {
		defer fmt.Println("defer", i)
	}

	// 規則2: 引数は defer 文を実行した時点で評価される
	x := 10
	defer fmt.Println("x at defer:", x)
	x = 20

	fmt.Println("x now:", x)
	fmt.Println("double:", double(5))
}

// 規則3: defer の関数は、名前付き戻り値を読み書きできる
func double(n int) (result int) {
	defer func() { result *= 2 }()
	return n // result = n がセットされた後に defer が実行される
}
```

```text
x now: 20
double: 10
x at defer: 10
defer 2
defer 1
defer 0
```

規則 2 を利用すると、関数の入口と出口を表示するトレース関数を短く書けます (Effective Go の例を元にしています)。

```go
package main

import "fmt"

func trace(s string) string {
	fmt.Println("entering:", s)
	return s
}

func un(s string) {
	fmt.Println("leaving:", s)
}

func a() {
	defer un(trace("a")) // trace("a") はこの時点で実行され、un は a の終了時に実行される
	fmt.Println("in a")
}

func b() {
	defer un(trace("b"))
	fmt.Println("in b")
	a()
}

func main() {
	b()
}
```

```text
entering: b
in b
entering: a
in a
leaving: a
leaving: b
```

> ⚠️ **ループの中で defer を使うときは注意**: defer が実行されるのは「ループの 1 周の終わり」ではなく「関数の終わり」です。大量のファイルをループで開いて `defer f.Close()` すると、関数を抜けるまでファイルが閉じられません。1 周分の処理を別の関数に切り出しましょう。

panic と recover での defer の使い方は第7章で扱います。

## 3.5 goto

Go には `goto` もありますが、使う場面はほとんどありません。生成されたコードや、ごく一部の低レベルな処理で見かける程度です。

## ✍️ 演習: ニュートン法で平方根

[`exercises/sqrt`](../exercises/sqrt/exercise.go) の `Sqrt` 関数を `for` ループで実装してください。

平方根 √x の近似値 z は、次の式を繰り返すと正しい値に近づいていきます (ニュートン法)。

```text
z -= (z*z - x) / (2*z)
```

z の初期値を 1.0 とし、**z の変化が 1e-10 未満になったら** ループを終えます。

```bash
cd exercises
go test ./sqrt/                   # テストが通れば正解
go test -tags solution ./sqrt/    # 模範解答で実行する場合
```

ヒント:

- `math.Abs` で絶対値を求められます。
- 何回の反復で収束したか、z の値を表示して観察してみましょう。x の値を変えると反復回数はどう変わるでしょうか。
- 解答例は [`exercises/sqrt/solution.go`](../exercises/sqrt/solution.go) にあります。
