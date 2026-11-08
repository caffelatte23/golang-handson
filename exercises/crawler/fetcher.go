// Package crawler は第8章の演習です。詳細は exercise.go を参照してください。
package crawler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Fetcher は URL の本文と、そのページに含まれるリンクを返します。
type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

// FakeFetcher はテスト用の Fetcher です (演習の対象ではありません)。
// 各 Fetch に Delay だけ時間がかかり、呼ばれた回数と最大同時実行数を記録します。
type FakeFetcher struct {
	Pages map[string]FakeResult
	Delay time.Duration

	mu      sync.Mutex
	Calls   map[string]int
	running atomic.Int32
	Peak    atomic.Int32
}

// FakeResult は 1 ページ分の結果です。
type FakeResult struct {
	Body string
	URLs []string
}

// Fetch は Fetcher を実装します。
func (f *FakeFetcher) Fetch(url string) (string, []string, error) {
	n := f.running.Add(1)
	defer f.running.Add(-1)
	for {
		p := f.Peak.Load()
		if n <= p || f.Peak.CompareAndSwap(p, n) {
			break
		}
	}

	f.mu.Lock()
	if f.Calls == nil {
		f.Calls = map[string]int{}
	}
	f.Calls[url]++
	f.mu.Unlock()

	time.Sleep(f.Delay)
	if res, ok := f.Pages[url]; ok {
		return res.Body, res.URLs, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}
