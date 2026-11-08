//go:build !solution

package crawler

// Crawl は url から depth 段までリンクを辿り、取得できたページの url → body を返します。
//
// 要件:
//   - 同じ URL を 2 回以上 Fetch しない
//   - 複数のページを並行に Fetch する (goroutine を使う)
//   - Fetch がエラーのページは結果に含めない
//   - go test -race で競合が出ない
//
// ヒント: 訪問済みの map を sync.Mutex で守り、sync.WaitGroup で全 goroutine の終了を待ちます。
//
// 確認: go test -race ./crawler/
func Crawl(url string, depth int, fetcher Fetcher) map[string]string {
	// TODO: 以下は逐次・重複ありの素朴な実装です。要件を満たすように書き換えてください。
	result := map[string]string{}
	var crawl func(string, int)
	crawl = func(url string, depth int) {
		if depth <= 0 {
			return
		}
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			return
		}
		result[url] = body
		for _, u := range urls {
			crawl(u, depth-1)
		}
	}
	crawl(url, depth)
	return result
}
