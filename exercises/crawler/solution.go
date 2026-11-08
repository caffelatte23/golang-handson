//go:build solution

package crawler

import "sync"

// Crawl は url から depth 段までリンクを辿り、取得できたページの url → body を返します。
func Crawl(url string, depth int, fetcher Fetcher) map[string]string {
	var (
		mu      sync.Mutex
		visited = map[string]bool{}
		result  = map[string]string{}
		wg      sync.WaitGroup
	)

	var crawl func(string, int)
	crawl = func(url string, depth int) {
		defer wg.Done()
		if depth <= 0 {
			return
		}

		// 「確認して登録」を 1 つのロックの中で行うことで、重複 Fetch を防ぐ
		mu.Lock()
		if visited[url] {
			mu.Unlock()
			return
		}
		visited[url] = true
		mu.Unlock()

		body, urls, err := fetcher.Fetch(url) // ロックの外で (並行に) 実行する
		if err != nil {
			return
		}

		mu.Lock()
		result[url] = body
		mu.Unlock()

		for _, u := range urls {
			wg.Add(1)
			go crawl(u, depth-1)
		}
	}

	wg.Add(1)
	go crawl(url, depth)
	wg.Wait()
	return result
}
