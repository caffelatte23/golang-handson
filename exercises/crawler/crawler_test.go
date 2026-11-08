package crawler

import (
	"testing"
	"time"
)

func newFetcher() *FakeFetcher {
	return &FakeFetcher{
		Delay: 20 * time.Millisecond,
		Pages: map[string]FakeResult{
			"https://golang.org/": {"The Go Programming Language", []string{
				"https://golang.org/pkg/", "https://golang.org/cmd/"}},
			"https://golang.org/pkg/": {"Packages", []string{
				"https://golang.org/", "https://golang.org/cmd/",
				"https://golang.org/pkg/fmt/", "https://golang.org/pkg/os/"}},
			"https://golang.org/pkg/fmt/": {"Package fmt", []string{
				"https://golang.org/", "https://golang.org/pkg/"}},
			"https://golang.org/pkg/os/": {"Package os", []string{
				"https://golang.org/", "https://golang.org/pkg/"}},
		},
	}
}

func TestCrawl(t *testing.T) {
	f := newFetcher()
	got := Crawl("https://golang.org/", 4, f)

	want := map[string]string{
		"https://golang.org/":         "The Go Programming Language",
		"https://golang.org/pkg/":     "Packages",
		"https://golang.org/pkg/fmt/": "Package fmt",
		"https://golang.org/pkg/os/":  "Package os",
	}
	if len(got) != len(want) {
		t.Errorf("got %d pages, want %d: %v", len(got), len(want), got)
	}
	for u, body := range want {
		if got[u] != body {
			t.Errorf("result[%s] = %q, want %q", u, got[u], body)
		}
	}
	for u, n := range f.Calls {
		if n > 1 {
			t.Errorf("%s fetched %d times, want at most once", u, n)
		}
	}
	if p := f.Peak.Load(); p < 2 {
		t.Errorf("peak concurrent fetches = %d, want >= 2 (fetch in parallel)", p)
	}
}

func TestCrawl_Depth(t *testing.T) {
	got := Crawl("https://golang.org/", 1, newFetcher())
	if len(got) != 1 {
		t.Errorf("depth 1 should fetch only the root: %v", got)
	}
}

// 深さの制限が経路の取り方に左右されないこと。
// root → slow → leaf (残り深さ 1 で leaf に到達)
// root → leaf → deep (残り深さ 2 で leaf に到達するので deep まで取得できる)
// DFS で先に slow 経由で leaf を「訪問済み」にすると、deep を取り逃がす。
func TestCrawl_DepthByShortestPath(t *testing.T) {
	f := &FakeFetcher{
		Pages: map[string]FakeResult{
			"root": {"root", []string{"slow", "leaf"}},
			"slow": {"slow", []string{"leaf"}},
			"leaf": {"leaf", []string{"deep"}},
			"deep": {"deep", nil},
		},
	}
	for range 20 { // 実行順に依存するバグを見つけやすくするため繰り返す
		got := Crawl("root", 3, f)
		if _, ok := got["deep"]; !ok || len(got) != 4 {
			t.Fatalf("got %v, want root, slow, leaf, deep", keys(got))
		}
	}
}

func keys(m map[string]string) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
