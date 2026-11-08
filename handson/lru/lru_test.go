package lru

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"
	"testing"
)

func TestCache_Eviction(t *testing.T) {
	c := New[string, int](2)
	var evicted []string
	c.OnEvict = func(k string, _ int) { evicted = append(evicted, k) }

	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")    // a が最新になる
	c.Put("c", 3) // 最も古い b が追い出される

	if _, ok := c.Get("b"); ok {
		t.Error("b should be evicted")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("Get(a) = %v, %v", v, ok)
	}
	if !slices.Equal(evicted, []string{"b"}) {
		t.Errorf("evicted = %v", evicted)
	}
}

func TestCache_All(t *testing.T) {
	c := New[int, string](3)
	for i := range 4 {
		c.Put(i, strconv.Itoa(i))
	}
	var keys []int
	for k := range c.All() {
		keys = append(keys, k)
	}
	if want := []int{3, 2, 1}; !slices.Equal(keys, want) {
		t.Errorf("keys = %v, want %v", keys, want)
	}
	// maps.Collect などイテレータを受け取る標準関数とも組み合わせられる
	if m := maps.Collect(c.All()); len(m) != 3 || m[2] != "2" {
		t.Errorf("collected = %v", m)
	}
}

func TestCache_Concurrent(t *testing.T) {
	c := New[int, int](100)
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 1000 {
				c.Put(i%200, g)
				c.Get(i % 50)
			}
		}()
	}
	wg.Wait()
	if c.Len() != 100 {
		t.Errorf("Len = %d", c.Len())
	}
}

// Example 関数は go doc / pkg.go.dev にサンプルとして表示され、
// // Output: コメントと出力が一致するかテストもされます。
func ExampleCache() {
	c := New[string, int](2)
	c.Put("apple", 100)
	c.Put("banana", 200)
	c.Put("cherry", 300) // apple が追い出される

	_, ok := c.Get("apple")
	fmt.Println("apple:", ok)
	for k, v := range c.All() {
		fmt.Println(k, v)
	}
	// Output:
	// apple: false
	// cherry 300
	// banana 200
}

// ベンチマーク: go test -bench=. -benchmem
func BenchmarkCache_Put(b *testing.B) {
	c := New[int, int](1024)
	i := 0
	for b.Loop() { // Go 1.24+
		c.Put(i%2048, i)
		i++
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := New[int, int](1024)
	for i := range 1024 {
		c.Put(i, i)
	}
	i := 0
	for b.Loop() {
		c.Get(i % 1024)
		i++
	}
}

// ファズテスト: go test -fuzz=FuzzCache -fuzztime=10s
// ランダムな操作列に対して「容量を超えない」「直前に Put した値が取れる」という性質を検証します。
func FuzzCache(f *testing.F) {
	f.Add(3, []byte{1, 2, 3, 1, 4, 5})
	f.Fuzz(func(t *testing.T, capacity int, ops []byte) {
		if capacity < 1 || capacity > 64 {
			t.Skip()
		}
		c := New[byte, int](capacity)
		for i, k := range ops {
			c.Put(k, i)
			if v, ok := c.Get(k); !ok || v != i {
				t.Fatalf("Get(%d) = %d, %v; want %d", k, v, ok, i)
			}
			if c.Len() > capacity {
				t.Fatalf("Len %d exceeds capacity %d", c.Len(), capacity)
			}
		}
	})
}
