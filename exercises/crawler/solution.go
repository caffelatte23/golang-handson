//go:build solution

package crawler

import "sync"

// Crawl は url から depth 段までリンクを辿り、取得できたページの url → body を返します。
//
// 1 段 (同じ深さの URL の集まり) ずつ処理する幅優先探索にしています。
// 各 URL には「最短の経路」で最初に到達するので、残りの深さが最大の状態で 1 回だけ取得でき、
// goroutine の実行順によって結果が変わりません。
func Crawl(url string, depth int, fetcher Fetcher) map[string]string {
	result := map[string]string{}
	visited := map[string]bool{url: true}
	frontier := []string{url}

	for ; depth > 0 && len(frontier) > 0; depth-- {
		var (
			mu   sync.Mutex
			wg   sync.WaitGroup
			next []string
		)
		for _, u := range frontier {
			wg.Add(1)
			go func() { // 同じ段の URL は並行に取得する
				defer wg.Done()
				body, urls, err := fetcher.Fetch(u)
				if err != nil {
					return
				}
				mu.Lock()
				defer mu.Unlock()
				result[u] = body
				for _, link := range urls {
					if !visited[link] { // 確認と登録を同じロックの中で行い、重複を防ぐ
						visited[link] = true
						next = append(next, link)
					}
				}
			}()
		}
		wg.Wait() // この段がすべて終わってから次の段へ進む
		frontier = next
	}
	return result
}
