# 第5章 並行処理

Go の最大の武器です。**goroutine** (軽量スレッド) と **channel** (goroutine 間の通信路) が言語に組み込まれています。

## 5.1 goroutine

```go
go doWork() // これだけで並行に実行される
```

goroutine は数 KB のスタックで始まり、ランタイムが OS スレッドに多重化するため、**数万〜数十万個** 起動しても問題ありません。

### 終了を待つ: `sync.WaitGroup`

```go
var wg sync.WaitGroup
for i := range 3 {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("worker", i)
	}()
}
wg.Wait()
```

> 💡 Go 1.25 では `wg.Go(func(){...})` という `Add`/`Done` を省略できるメソッドが追加されました。

## 5.2 channel

```go
ch := make(chan int)      // バッファなし: 送信と受信が同期する
buf := make(chan int, 10) // バッファあり

go func() {
	for i := range 5 {
		ch <- i // 送信
	}
	close(ch) // もう送らないことを通知 (送信側が close する)
}()

for v := range ch { // close されるまで受信
	fmt.Println(v)
}
```

### select — 複数チャネルの待ち合わせ

```go
select {
case msg := <-msgs:
	handle(msg)
case <-time.After(3 * time.Second):
	fmt.Println("timeout")
case <-ctx.Done():
	return ctx.Err()
}
```

## 5.3 共有メモリと mutex

channel が万能というわけではありません。**状態を守るだけなら mutex のほうが単純** なことも多いです。

```go
type Counter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n[key]++
}
```

| 使い分け | 道具 |
| --- | --- |
| データの所有権を渡す、処理をパイプラインでつなぐ、イベント通知 | channel |
| キャッシュやカウンタなど共有状態の保護 | `sync.Mutex` / `sync.RWMutex` |
| 単純なカウンタやフラグ | `sync/atomic` (`atomic.Int64` など) |
| 一度だけ初期化 | `sync.Once` / `sync.OnceValue` |

## 5.4 context — キャンセルとタイムアウトの伝播

`context.Context` は「この処理はいつまで続けてよいか」を関数呼び出しの連鎖に伝える仕組みです。

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // 必ず呼ぶ (リソースリーク防止)

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := http.DefaultClient.Do(req) // 2 秒を超えると中断される
```

**約束事**:

- 第 1 引数に `ctx context.Context` として渡す (構造体のフィールドに保存しない)
- HTTP ハンドラでは `r.Context()` を下位へ渡す。クライアントが切断するとキャンセルされる
- `Ctrl+C` で全体を止めたいときは `signal.NotifyContext`

## 5.5 errgroup — 「並行に実行して、どれかが失敗したら全部止める」

`golang.org/x/sync/errgroup` は、OSS で非常によく使われるパッケージです。

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // 同時実行数の上限 (ワーカープール相当)

for _, url := range urls {
	g.Go(func() error {
		return fetch(ctx, url) // 1 つでもエラーを返すと ctx がキャンセルされる
	})
}
if err := g.Wait(); err != nil { // 最初のエラーが返る
	return err
}
```

> 📦 **OSS ではこう書かれている**: 2026 年 11 月時点で、Kubernetes、Hugo、GitHub CLI、Prometheus、Caddy、containerd、Gitea、k6、MinIO、lazygit、Traefik の `go.mod` すべてに `golang.org/x/sync` が含まれていました。事実上の準標準ライブラリです。

## 5.6 よくあるパターン

### ワーカープール (channel 版)

```go
jobs := make(chan Job)
results := make(chan Result)

var wg sync.WaitGroup
for range numWorkers {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs {
			results <- process(j)
		}
	}()
}

go func() { wg.Wait(); close(results) }() // 全ワーカー終了で results を閉じる

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

### パイプライン / ファンアウト・ファンイン

`生成 → 変換 → 集約` を channel でつなぐパターン。詳細は Go Blog の [Pipelines and cancellation](https://go.dev/blog/pipelines) が必読です。

## 5.7 落とし穴と対策

| 問題 | 症状 | 対策 |
| --- | --- | --- |
| データ競合 | たまに値がおかしい、クラッシュ | `go test -race` / `go run -race` で検出。CI で常時有効に |
| goroutine リーク | メモリが増え続ける | 終了条件 (ctx, close) を必ず用意。テストで [`goleak`](https://github.com/uber-go/goleak) |
| close 済みチャネルへの送信 | panic | 「送信側が close する」原則を守る |
| nil map への書き込み | panic | `make` で初期化 |
| `defer` をループ内で使う | ループ終了までリソースが解放されない | 関数に切り出す |

## ✍️ 練習問題

1. 1〜100 の数値を 4 つのワーカーで 2 乗し、合計を求めよう (channel 版と errgroup 版の両方)。
2. `go test -race` でデータ競合を検出するテストをわざと書いてみよう (mutex なしで map に書き込む)。

→ 第12章のハンズオンでは、errgroup と context を使って **並行リンクチェッカー** を作ります。
