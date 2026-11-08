# 第8章 並行処理

並行処理は Go の最大の特徴です。**goroutine** (軽量なスレッド) と **channel** (goroutine 同士の通信路) が言語そのものに組み込まれています。

## 8.1 通信によってメモリを共有する

> *Do not communicate by sharing memory; instead, share memory by communicating.*
> (メモリを共有することで通信するのではなく、通信することでメモリを共有せよ)

多くの言語では、複数のスレッドが同じ変数を読み書きし、それをロックで保護します。Go が推奨するのは別の考え方で、**データそのものを channel で受け渡し、その時点でデータを触ってよいのは 1 つの goroutine だけ** という状態を保ちます。こうすると、設計の段階でデータ競合が起きなくなります。

もちろんロック (`sync.Mutex`) も使えます。どちらを選ぶかは 8.5 節で整理します。

## 8.2 goroutine

関数呼び出しの前に `go` を付けると、その関数は **新しい goroutine で並行に実行** されます。

```go
package main

import (
	"fmt"
	"sync"
)

func say(s string, wg *sync.WaitGroup) {
	defer wg.Done() // 終わったらカウンタを 1 減らす
	fmt.Println(s)
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2) // 待つ goroutine の数
	go say("world", &wg)
	go say("hello", &wg)
	wg.Wait() // カウンタが 0 になるまで待つ
	fmt.Println("done")
}
```

<!-- 出力順は実行ごとに変わる -->
```text
hello
world
done
```

(hello と world の順番は実行するたびに変わる可能性があります。)

- goroutine は数 KB の小さなスタックで始まり、必要に応じて伸び縮みします。Go ランタイムが多数の goroutine を少数の OS スレッドに割り当てるので、**数万〜数十万個を同時に動かしても問題ありません**。
- **`main` 関数が終わると、他の goroutine が途中でもプログラムは終了します**。終了を待ちたいときは `sync.WaitGroup` や channel を使います。
- Go 1.25 以降は、`wg.Go(func() { ... })` と書くと `Add(1)` と `defer Done()` を省略できます。

関数リテラルを使って、その場で goroutine を起動することもよくあります。

```go
func Announce(message string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		fmt.Println(message)
	}() // 末尾の () を忘れずに。関数を定義して、すぐに呼び出している
}
```

## 8.3 channel

channel は、値を送受信するための型付きの通信路です。演算子 `<-` で送信と受信を行います。

```go
ch <- v   // v を ch に送信する
v := <-ch // ch から受信して v に代入する
```

```go
package main

import "fmt"

func sum(s []int, c chan<- int) { // chan<- は送信専用の channel 型
	total := 0
	for _, v := range s {
		total += v
	}
	c <- total
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c1, c2 := make(chan int), make(chan int)
	go sum(s[:len(s)/2], c1) // 前半の合計
	go sum(s[len(s)/2:], c2) // 後半の合計
	x, y := <-c1, <-c2       // 両方の結果を受信するまで待つ

	fmt.Println(x, y, x+y)
}
```

```text
17 -5 12
```

- **バッファなしの channel では、送信と受信の両方の準備ができるまで、どちらも待たされます**。そのため、ロックや条件変数を使わなくても goroutine 同士の同期がとれます。
- `chan<- T` は送信専用、`<-chan T` は受信専用の channel 型です。関数の引数をこれらの型にすると、使い方の誤りをコンパイル時に防げます。

### バッファ付き channel

```go
package main

import "fmt"

func main() {
	ch := make(chan int, 2) // バッファの大きさは 2
	ch <- 1
	ch <- 2 // バッファに空きがあるうちは、送信してもブロックしない
	// ch <- 3 // ここで送信すると、受信する goroutine がいないので永久に待つ (deadlock)
	fmt.Println(<-ch, <-ch, len(ch), cap(ch))
}
```

```text
1 2 0 2
```

すべての goroutine が待ち状態になると、ランタイムが `fatal error: all goroutines are asleep - deadlock!` を表示して終了します。

### close と range

送信する側は、**もう値を送らないこと** を `close` で受信側に知らせられます。

```go
package main

import "fmt"

func fibonacci(n int, c chan<- int) {
	x, y := 0, 1
	for range n {
		c <- x
		x, y = y, x+y
	}
	close(c)
}

func main() {
	c := make(chan int, 10)
	go fibonacci(cap(c), c)
	for v := range c { // close されるまで受信を続ける
		fmt.Print(v, " ")
	}
	fmt.Println()

	v, ok := <-c // close された channel から受信すると、ゼロ値と false が返る
	fmt.Println(v, ok)
}
```

```text
0 1 1 2 3 5 8 13 21 34 
0 false
```

- **close するのは送信側** です。close された channel に送信すると panic します。
- channel はファイルとは違い、**必ずしも close する必要はありません**。受信側に「終わり」を伝える必要があるとき (`range` で受信しているときなど) だけ close します。

## 8.4 select

`select` は、複数の channel 操作のうち **準備ができたもの** を 1 つ選んで実行します。複数の準備ができている場合は、ランダムに 1 つを選びます。

```go
package main

import "fmt"

func fibonacci(c chan<- int, quit <-chan struct{}) {
	x, y := 0, 1
	for {
		select {
		case c <- x: // 受信側が受け取れる状態なら送信する
			x, y = y, x+y
		case <-quit: // quit から通知が来たら終わる
			fmt.Println("quit")
			return
		}
	}
}

func main() {
	c := make(chan int)
	quit := make(chan struct{}) // 中身のない struct{} は「通知だけ」を表す
	go func() {
		for range 10 {
			fmt.Print(<-c, " ")
		}
		fmt.Println()
		close(quit) // close は、受信しているすべての goroutine への一斉通知になる
	}()
	fibonacci(c, quit)
}
```

```text
0 1 1 2 3 5 8 13 21 34 
quit
```

### default とタイムアウト

```go
select {
case msg := <-msgs:
	handle(msg)
case <-time.After(3 * time.Second): // 3 秒待っても来なければタイムアウト
	return errors.New("timeout")
case <-ctx.Done(): // キャンセルされた
	return ctx.Err()
default:
	// どの channel も準備ができていなければ、待たずにすぐここが実行される (ノンブロッキング)
}
```

## 8.5 sync パッケージ — 共有メモリを守る

channel が常に最適とは限りません。**複数の goroutine から 1 つの値を安全に読み書きしたいだけなら、ロックのほうが簡単** です。

```go
package main

import (
	"fmt"
	"sync"
)

// SafeCounter は並行に使っても安全なカウンタ
type SafeCounter struct {
	mu sync.Mutex // 守りたいフィールドのすぐ上に置くのが慣習
	v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v[key]++
}

func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v[key]
}

func main() {
	c := SafeCounter{v: make(map[string]int)}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc("somekey")
		}()
	}
	wg.Wait()
	fmt.Println(c.Value("somekey"))
}
```

```text
1000
```

`mu` を外して `go run -race .` で実行すると、データ競合が報告されます。マップへの同時書き込みは、`-race` を付けなくてもランタイムが `fatal error: concurrent map writes` を出して停止させることがあります。

### 道具の使い分け

| やりたいこと | 道具 |
| --- | --- |
| データの所有権を渡す、処理をつないでパイプラインにする、イベントを通知する | channel |
| キャッシュやカウンタなど、共有している状態を守る | `sync.Mutex` / `sync.RWMutex` (読み取りが多いとき) |
| 単純なカウンタやフラグ | `sync/atomic` (`atomic.Int64`、`atomic.Bool` など) |
| 一度だけ初期化する | `sync.Once` / `sync.OnceValue` (Go 1.21+) |
| 複数の goroutine の終了を待つ | `sync.WaitGroup` |
| 一時的なオブジェクトを使い回して GC の負荷を減らす | `sync.Pool` |

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var loadConfig = sync.OnceValue(func() map[string]string {
	fmt.Println("loading config...") // 何回呼ばれても 1 回しか実行されない
	return map[string]string{"env": "prod"}
})

func main() {
	var hits atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = loadConfig()["env"]
			hits.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println(hits.Load())
}
```

```text
loading config...
100
```

## 8.6 Effective Go の並行処理パターン

### バッファ付き channel をセマフォとして使う

同時に実行する数を制限したいときは、バッファ付き channel を「空き枠の数」として使えます。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

const maxOutstanding = 3

func main() {
	sem := make(chan struct{}, maxOutstanding)
	var mu sync.Mutex
	running, peak := 0, 0
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // 枠を 1 つ取る (空きがなければ待つ)
			defer func() { <-sem }() // 終わったら枠を返す

			mu.Lock()
			running++
			peak = max(peak, running) // 同時に実行している数の最大値を記録する
			mu.Unlock()

			time.Sleep(10 * time.Millisecond) // 実際の処理の代わり

			mu.Lock()
			running--
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("max concurrent:", peak)
}
```

```text
max concurrent: 3
```

実際の開発では、同じことを `golang.org/x/sync/semaphore` や `errgroup.SetLimit` (8.8 節) を使って書くことが多いです。

### channel の channel — 返信用の channel を一緒に送る

リクエストの中に **返信用の channel** を入れて送ると、クライアントごとに結果を受け取れるサーバーが簡単に作れます。

```go
package main

import "fmt"

type Request struct {
	args       []int
	resultChan chan int // 結果はここに返してもらう
}

func sum(a []int) (s int) {
	for _, v := range a {
		s += v
	}
	return
}

func handle(queue <-chan *Request) {
	for req := range queue {
		req.resultChan <- sum(req.args)
	}
}

func main() {
	queue := make(chan *Request)
	for range 3 { // ワーカーを 3 つ起動する
		go handle(queue)
	}

	reqs := []*Request{
		{[]int{3, 4, 5}, make(chan int)},
		{[]int{10, 20}, make(chan int)},
	}
	for _, r := range reqs {
		go func() { queue <- r }()
	}
	for _, r := range reqs {
		fmt.Println(r.args, "=", <-r.resultChan)
	}
}
```

```text
[3 4 5] = 12
[10 20] = 30
```

### 並列化 — CPU コアに処理を分ける

計算を独立した部分に分けられるなら、CPU のコア数だけ goroutine を起動して並列に計算できます。

```go
package main

import (
	"fmt"
	"runtime"
)

func sumSquares(xs []float64, c chan<- float64) {
	var s float64
	for _, x := range xs {
		s += x * x
	}
	c <- s
}

func main() {
	xs := make([]float64, 1_000_000)
	for i := range xs {
		xs[i] = 1
	}

	numCPU := runtime.NumCPU()
	c := make(chan float64, numCPU) // バッファを用意しておけば、送信側は待たされない
	chunk := (len(xs) + numCPU - 1) / numCPU
	for i := 0; i < len(xs); i += chunk {
		go sumSquares(xs[i:min(i+chunk, len(xs))], c)
	}

	var total float64
	for i := 0; i < len(xs); i += chunk {
		total += <-c
	}
	fmt.Println(total)
}
```

```text
1e+06
```

> 💡 **並行 (concurrency) と並列 (parallelism) は別の概念です**。並行とは、独立して動く処理の組み合わせとしてプログラムを **構成する** ことです。並列とは、複数の計算を **同時に実行** して速くすることです。Go は並行に書くための言語であり、並列に実行されるかどうかはその結果です (Rob Pike の講演 [Concurrency is not Parallelism](https://go.dev/blog/waza-talk))。

### リークするバッファ (free list)

`select` の `default` を使うと、「空いていれば使い、なければ新しく作る」「満杯なら捨てる」という処理をノンブロッキングで書けます。Effective Go では、バッファを再利用するフリーリストの例として紹介されています。

```go
var freeList = make(chan *Buffer, 100)

func getBuffer() *Buffer {
	select {
	case b := <-freeList:
		return b // 再利用できるバッファがあれば使う
	default:
		return new(Buffer) // なければ新しく確保する
	}
}

func putBuffer(b *Buffer) {
	select {
	case freeList <- b: // フリーリストに空きがあれば戻す
	default: // 満杯なら何もしない (GC が回収する = ここで "リーク" させる)
	}
}
```

現在では、同じ目的に `sync.Pool` が使えます。

## 8.7 context — キャンセルとタイムアウトを伝える

`context.Context` は、**この処理をいつまで続けてよいか** を、関数呼び出しの連鎖を通じて伝える仕組みです。HTTP サーバーや RPC、データベースへのアクセスで必須になります。

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func slowOperation(ctx context.Context) error {
	select {
	case <-time.After(1 * time.Second): // 1 秒かかる処理
		return nil
	case <-ctx.Done(): // それより先にキャンセルされたら中断する
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel() // タイムアウトする前に処理が終わっても、必ず呼んでリソースを解放する

	err := slowOperation(ctx)
	fmt.Println(err, errors.Is(err, context.DeadlineExceeded))
}
```

```text
context deadline exceeded true
```

**context を使うときの約束事**

- **関数の第 1 引数** に `ctx context.Context` という名前で渡します。構造体のフィールドに保存してはいけません。
- HTTP ハンドラでは `r.Context()` を下の層に渡します。クライアントが接続を切ると、自動的にキャンセルされます。
- `Ctrl+C` でプログラム全体を止めたいときは、`signal.NotifyContext` を使います (第15章)。
- `context.WithValue` に入れるのは、リクエスト ID のような **リクエストの範囲だけで使うデータ** に限ります。関数の省略可能な引数の代わりに使ってはいけません。

## 8.8 errgroup — 並行に実行し、1 つでも失敗したら全体を止める

`golang.org/x/sync/errgroup` は、OSS で非常によく使われているパッケージです。

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // 同時に実行する数の上限 (ワーカープールと同じ効果)

for _, url := range urls {
	g.Go(func() error {
		return fetch(ctx, url) // どれか 1 つがエラーを返すと、ctx がキャンセルされる
	})
}
if err := g.Wait(); err != nil { // 全部の終了を待ち、最初に起きたエラーを返す
	return err
}
```

> 📦 **OSS ではこう書かれている**: 2026 年 11 月時点で、Kubernetes、Hugo、GitHub CLI、Prometheus、Caddy、containerd、Gitea、k6、MinIO、lazygit、Traefik の `go.mod` すべてが `golang.org/x/sync` に直接依存していました。事実上、標準ライブラリに準ずる扱いです。

第16章のハンズオンでは、errgroup と context を使って並行リンクチェッカーを作ります。

## 8.9 よく使うパターン

### ワーカープール (channel で書く場合)

```go
jobs := make(chan Job)
results := make(chan Result)

var wg sync.WaitGroup
for range numWorkers {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs { // jobs が close されるまで処理を続ける
			results <- process(j)
		}
	}()
}

go func() { wg.Wait(); close(results) }() // 全ワーカーが終わったら results を close する

go func() {
	for _, j := range allJobs {
		jobs <- j
	}
	close(jobs)
}()

for r := range results {
	fmt.Println(r)
}
```

### パイプライン

「生成 → 変換 → 集約」の各段階を goroutine にし、channel でつなぎます。

```go
package main

import "fmt"

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	for v := range sq(sq(gen(1, 2, 3))) { // 段階を自由に組み合わせられる
		fmt.Println(v)
	}
}
```

```text
1
16
81
```

途中で処理をやめるときにキャンセルを伝える方法などは、Go Blog の [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) が詳しいです。

## 8.10 落とし穴と対策

| 問題 | 症状 | 対策 |
| --- | --- | --- |
| データ競合 | ときどき値がおかしくなる、クラッシュする | `go test -race` / `go run -race` で検出できる。CI では常に有効にする |
| goroutine リーク | メモリ使用量が増え続ける | 終了条件 (ctx のキャンセル、channel の close) を必ず用意する。テストでは [`goleak`](https://github.com/uber-go/goleak) で検出できる |
| deadlock | プログラムが止まる | 送信と受信の数を合わせる。待つ処理には `select` と ctx でタイムアウトを付ける |
| close 済みの channel への送信 | panic | 「close するのは送信側」の原則を守る |
| `WaitGroup.Add` を goroutine の中で呼ぶ | `Wait` が早く戻ってしまう | `Add` は `go` 文の前に呼ぶ (または Go 1.25 の `wg.Go` を使う) |
| `sync.Mutex` を含む構造体のコピー | ロックが効かない | ポインタで渡す。`go vet` の copylocks チェックで検出できる |

## ✍️ 演習

### 演習 8-1: 同じ値を持つ二分木

二分探索木は、同じ値の集合でも形がいろいろあり得ます。[`exercises/equivtree`](../exercises/equivtree/exercise.go) で、2 つの木が同じ値の列を持つかどうかを goroutine と channel を使って判定してください。

1. `Walk(t *tree.Tree, ch chan int)`: 木の値を小さい順に `ch` へ送り、最後に `close(ch)` します。
2. `Same(t1, t2 *tree.Tree) bool`: 2 つの木を並行に `Walk` しながら値を比べます。

`tree.New(k)` は、k, 2k, ..., 10k を値に持つ、ランダムな形の木を返します。

```bash
cd exercises
go test -race ./equivtree/
```

考えてみよう: `Same` で値が違うと分かった時点で return すると、`Walk` の goroutine はどうなるでしょうか (解答例ではリークを防ぐ処理をしています)。

### 演習 8-2: 並行 Web クローラー

[`exercises/crawler`](../exercises/crawler/exercise.go) の `Crawl` は、URL を 1 つずつ順に取得し、同じ URL も何度も取得してしまいます。次の要件を満たすように書き換えてください。

- 同じ URL は 1 回しか取得しない
- 複数のページを並行に取得する
- `go test -race` で競合が報告されない

```bash
go test -race ./crawler/
```

ヒント: 訪問済みの URL を記録するマップを `sync.Mutex` で守り、`sync.WaitGroup` ですべての goroutine の終了を待ちます。
