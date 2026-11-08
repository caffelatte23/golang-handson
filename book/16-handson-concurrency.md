# 第16章 ハンズオン③ 並行処理 — リンクチェッカー (goroutine + errgroup)

Go の並行処理が最も活きるのは **I/O 待ちが多い処理** です。この章では、大量の URL を並行にチェックするコマンドを作り、goroutine・context・errgroup の実践的な使い方を学びます。

📁 完成コード: [`handson/linkcheck/`](../handson/linkcheck/)

## 学べること

- `errgroup` + `SetLimit` による **同時実行数を制限した並行処理**
- `context` によるタイムアウトと Ctrl+C でのキャンセル
- 結果の順序を保つテクニック (インデックス書き込み)
- 標準 `flag` パッケージだけで作る CLI と、標準入力からの読み込み (Unix パイプライン対応)
- `run(args, stdin, stdout, stderr) int` パターンによる main のテスト
- 並行処理のテスト (同時実行数の検証・キャンセルの検証)

## 完成イメージ

```console
$ linkcheck https://go.dev https://go.dev/no-such-page
✓ 200  https://go.dev                                     152ms
✗ 404  https://go.dev/no-such-page                        98ms

2 checked, 1 broken

$ cat urls.txt | linkcheck -c 16 -timeout 30s
```

終了コード: `0` = 全て正常、`1` = リンク切れあり、`2` = 使い方の誤り・中断。CI に組み込みやすい設計です。

## Step 1. プロジェクトを作る

```bash
mkdir linkcheck && cd linkcheck
go mod init github.com/<your-name>/linkcheck
go get golang.org/x/sync/errgroup
mkdir -p internal/checker
```

## Step 2. まずは逐次版で考える

いきなり並行にせず、まず 1 件ずつ処理する素朴な版を考えます。

```go
for _, u := range urls {
	results = append(results, checkOne(ctx, u))
}
```

100 件 × 平均 300ms なら 30 秒かかります。ネットワーク待ちの間 CPU は暇なので、並行化すれば劇的に速くなります。

## Step 3. 並行版のチェッカー

`internal/checker/checker.go`:

```go
// handson/linkcheck/internal/checker/checker.go
// Package checker は複数の URL を並行にチェックします。
package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// Result は 1 件のチェック結果です。
type Result struct {
	URL      string
	Status   int
	Duration time.Duration
	Err      error
}

// OK はリンクが正常かどうかを返します (ステータスが 400 未満 = 2xx/3xx なら正常)。
func (r Result) OK() bool { return r.Err == nil && r.Status < 400 }

// Checker は設定を保持します。
type Checker struct {
	Client      *http.Client
	Concurrency int // 同時実行数の上限
}

// Check は urls を並行にチェックし、入力と同じ順序で結果を返します。
//
// ポイント:
//   - errgroup.SetLimit で goroutine の同時実行数を制限する (ワーカープール相当)
//   - 結果は「インデックス指定でスライスに書き込む」ので mutex 不要
//   - ctx がキャンセルされると、実行中のリクエストも中断される
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	results := make([]Result, len(urls))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(max(c.Concurrency, 1))

	for i, u := range urls {
		g.Go(func() error { // Go 1.22+ ではループ変数がイテレーションごとに別物なので安全
			if err := ctx.Err(); err != nil {
				return err // すでに中断されていれば実行しない (results[i] はゼロ値のまま)
			}
			results[i] = c.checkOne(ctx, u)
			// 個々のリンク切れは「結果」であってエラーではないので nil を返す。
			// ctx のキャンセルだけは全体を止めるべきエラーとして返す。
			return ctx.Err()
		})
	}
	if err := g.Wait(); err != nil {
		return results, fmt.Errorf("check aborted: %w", err)
	}
	return results, nil
}

func (c *Checker) checkOne(ctx context.Context, url string) Result {
	start := time.Now()
	res := Result{URL: url}

	// まず軽い HEAD で確認し、HEAD に対応していないサーバー (405/501) なら GET で再試行する
	status, err := c.do(ctx, http.MethodHead, url)
	if err == nil && (status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented) {
		status, err = c.do(ctx, http.MethodGet, url)
	}
	res.Status, res.Err, res.Duration = status, err, time.Since(start)
	return res
}

func (c *Checker) do(ctx context.Context, method, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "linkcheck-handson/1.0")

	resp, err := c.Client.Do(req)
	if err != nil {
		return 0, err
	}
	// Body は必ず Close する (bodyclose linter がチェック)。
	// GET の本文は読まないので、すぐに閉じる
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
```

**ポイント**

1. **同時実行数の制限**: `g.SetLimit(n)` を設定すると、`g.Go` は実行中の goroutine が n 個未満になるまでブロックします。無制限に goroutine を起動すると、相手サーバーへの負荷やファイルディスクリプタ枯渇を招きます。
2. **結果の順序保持**: `results[i] = ...` のように **各 goroutine が別々のインデックスに書く** だけならデータ競合にならず、mutex も channel も不要です。
3. **ループ変数のキャプチャ**: Go 1.22 以降 (`go.mod` の `go` 行が 1.22 以上) は `i`, `u` がイテレーションごとに新しい変数になるので、そのままクロージャで使えます。
4. **エラーの扱い分け**: 「404 だった」はチェック結果であり、全体を止める理由ではありません。errgroup にエラーとして返すのは「ctx がキャンセルされた」ときだけです。
5. **HEAD → GET のフォールバック**: 本文を取得しない軽い `HEAD` リクエストで確認しますが、HEAD に対応していないサーバーは 405 や 501 を返すため、その場合だけ `GET` で再試行します。
6. **`defer resp.Body.Close()`**: レスポンスのボディを閉じ忘れると、コネクションが再利用されずリークします。

### channel 版との比較

同じことは第8章のワーカープール (jobs/results チャネル) でも書けます。errgroup 版のほうが、

- エラー伝播とキャンセルが組み込み
- `WaitGroup` や `close` の管理が不要

と短く安全なため、**OSS ではこちらが主流** です (第8章参照: 調査対象 15 中 11 リポジトリが `golang.org/x/sync` に直接依存)。

## Step 4. 並行処理のテスト

`internal/checker/checker_test.go`:

```go
// handson/linkcheck/internal/checker/checker_test.go
package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestChecker_Check(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/missing", http.NotFound)
	mux.HandleFunc("/get-only", func(w http.ResponseWriter, r *http.Request) { // HEAD 非対応のサーバーを模す
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Checker{Client: srv.Client(), Concurrency: 2}
	urls := []string{srv.URL + "/ok", srv.URL + "/missing", "http://127.0.0.1:1/unreachable", srv.URL + "/get-only"}

	got, err := c.Check(context.Background(), urls)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(urls) {
		t.Fatalf("len = %d", len(got))
	}
	// 入力順が保たれていること
	for i := range urls {
		if got[i].URL != urls[i] {
			t.Errorf("got[%d].URL = %s, want %s", i, got[i].URL, urls[i])
		}
	}
	if !got[0].OK() {
		t.Errorf("/ok should be OK: %+v", got[0])
	}
	if got[1].OK() || got[1].Status != http.StatusNotFound {
		t.Errorf("/missing should be 404: %+v", got[1])
	}
	if got[2].Err == nil {
		t.Errorf("unreachable should have error")
	}
	if !got[3].OK() {
		t.Errorf("/get-only should be OK via GET fallback: %+v", got[3])
	}
}

func TestChecker_ConcurrencyLimit(t *testing.T) {
	var cur, peak atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := cur.Add(1)
		defer cur.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}))
	defer srv.Close()

	urls := make([]string, 20)
	for i := range urls {
		urls[i] = srv.URL
	}
	c := &Checker{Client: srv.Client(), Concurrency: 3}
	if _, err := c.Check(context.Background(), urls); err != nil {
		t.Fatal(err)
	}
	if p := peak.Load(); p > 3 {
		t.Errorf("peak concurrency = %d, want <= 3", p)
	}
}

func TestChecker_Cancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := &Checker{Client: srv.Client(), Concurrency: 4}
	start := time.Now()
	_, err := c.Check(ctx, []string{srv.URL, srv.URL, srv.URL})
	if err == nil {
		t.Fatal("want error on timeout")
	}
	if time.Since(start) > time.Second {
		t.Errorf("cancellation took too long: %v", time.Since(start))
	}
}
```

**ポイント**

- **外部ネットワークに依存しない**: `httptest.NewServer` でローカルにテスト用サーバーを立てます。
- **同時実行数の検証**: サーバー側で `atomic.Int32` を使って同時接続数のピークを記録し、`SetLimit` が効いていることを確かめます。
- **キャンセルの検証**: 応答しないサーバーに対して 50ms のタイムアウトを設定し、すぐに戻ってくることを確認します。
- 必ず `-race` 付きで実行しましょう。

```bash
go test -race -v ./internal/checker/
```

## Step 5. CLI 部分 (標準 flag)

`main.go`:

```go
// handson/linkcheck/main.go
// linkcheck は URL のリストを並行にチェックする CLI です。
//
//	linkcheck https://go.dev https://example.com/404
//	cat urls.txt | linkcheck -c 16 -timeout 3s
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/caffelatte23/golang-hanson/handson/linkcheck/internal/checker"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// 標準ライブラリ flag だけでも小さな CLI は十分書ける
	fs := flag.NewFlagSet("linkcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	concurrency := fs.Int("c", 8, "同時実行数")
	timeout := fs.Duration("timeout", 10*time.Second, "全体のタイムアウト")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	urls := fs.Args()
	if len(urls) == 0 { // 引数がなければ標準入力から 1 行 1 URL で読む
		sc := bufio.NewScanner(stdin)
		for sc.Scan() {
			if line := strings.TrimSpace(sc.Text()); line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintln(stderr, "read stdin:", err)
			return 2
		}
	}
	if len(urls) == 0 {
		fs.Usage()
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, *timeout)
	defer cancel()

	c := &checker.Checker{
		Client:      &http.Client{Timeout: 5 * time.Second},
		Concurrency: *concurrency,
	}
	results, err := c.Check(ctx, urls)

	broken := 0
	for _, r := range results {
		switch {
		case r.URL == "":
			continue // 中断されたため実行されなかった URL
		case r.Err != nil:
			broken++
			fmt.Fprintf(stdout, "✗ ERR  %-50s %v\n", r.URL, r.Err)
		case !r.OK():
			broken++
			fmt.Fprintf(stdout, "✗ %d  %-50s %v\n", r.Status, r.URL, r.Duration.Round(time.Millisecond))
		default:
			fmt.Fprintf(stdout, "✓ %d  %-50s %v\n", r.Status, r.URL, r.Duration.Round(time.Millisecond))
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	fmt.Fprintf(stdout, "\n%d checked, %d broken\n", len(results), broken)
	if broken > 0 {
		return 1
	}
	return 0
}
```

**ポイント**

- **`run(args, stdin, stdout, stderr) int` パターン**: `main` から OS 依存部分 (`os.Args`, `os.Stdin`, `os.Exit`) を切り離すと、`main_test.go` から普通の関数としてテストできます。
- `flag.NewFlagSet(..., flag.ContinueOnError)` を使うと、パース失敗時に `os.Exit` されずエラーが返ります (テストしやすい)。
- 引数がなければ標準入力を読むので、`grep -o 'https://[^)]*' README.md | linkcheck` のように **Unix パイプラインに組み込める** CLI になります。
- `signal.NotifyContext` と `context.WithTimeout` を **重ねて** 使い、Ctrl+C とタイムアウトのどちらでも止まるようにしています。

`main_test.go` (抜粋):

```go
func TestRun(t *testing.T) {
	// ...
	t.Run("broken via stdin", func(t *testing.T) {
		var out, errOut bytes.Buffer
		in := strings.NewReader("# comment\n" + srv.URL + "/ok\n" + srv.URL + "/nope\n")
		code := run(nil, in, &out, &errOut)
		if code != 1 || !strings.Contains(out.String(), "2 checked, 1 broken") {
			t.Fatalf("code = %d, out=%s", code, &out)
		}
	})
}
```

## Step 6. 動かしてみる

```bash
go build -o linkcheck .
./linkcheck https://go.dev https://pkg.go.dev https://go.dev/no-such-page

# このリポジトリの Markdown 中のリンクを全部チェック
grep -rhoE 'https://[^) ">]+' ../../book | sort -u | ./linkcheck -c 16
```

`-c 1` と `-c 16` で実行時間を比べてみると、並行化の効果を体感できます。

```bash
time (grep -rhoE 'https://[^) ">]+' ../../book | sort -u | ./linkcheck -c 1  > /dev/null)
time (grep -rhoE 'https://[^) ">]+' ../../book | sort -u | ./linkcheck -c 16 > /dev/null)
```

## 🚀 発展課題

1. **ホストごとのレート制限**: 同じホストへの同時接続を 2 本までにしよう (`map[string]*semaphore.Weighted` や `golang.org/x/time/rate`)。
2. **リトライ**: 5xx やタイムアウト時に指数バックオフで最大 3 回リトライしよう。
3. **HEAD を拒否するサーバー**: 403 を返すサーバーもあります。403 のときも GET で再試行するべきか、誤検知と見逃しのトレードオフを考えて実装しよう。
4. **クローラー化**: HTML をパースして (`golang.org/x/net/html`) 同一ドメインのリンクを再帰的に辿ろう。訪問済み管理には mutex 付き map、深さ制限も忘れずに。
5. **進捗表示**: 処理済み件数を stderr にリアルタイム表示しよう (`atomic.Int64` + `time.Ticker`)。
6. **プロファイリング**: `go test -cpuprofile` や `net/http/pprof` でボトルネックを調べてみよう。
