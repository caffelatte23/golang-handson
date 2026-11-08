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
