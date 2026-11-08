// Package lru はジェネリクスを使ったスレッドセーフな LRU キャッシュです。
//
// 最も長く使われていない要素から追い出します。
//
//	c := lru.New[string, int](2)
//	c.Put("a", 1)
//	v, ok := c.Get("a")
package lru

import (
	"container/list"
	"iter"
	"sync"
)

// entry はリストの要素として保持するキーと値の組です。
type entry[K comparable, V any] struct {
	key   K
	value V
}

// Cache は容量固定の LRU キャッシュです。ゼロ値は使えないので New で作成します。
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List // 先頭ほど最近使われた要素
	items    map[K]*list.Element

	// OnEvict が設定されていれば、追い出し時に呼ばれます。
	//
	// 注意:
	//   - キャッシュのロックを保持したまま呼ばれるため、OnEvict の中から
	//     同じ Cache のメソッドを呼ぶとデッドロックします。
	//   - New の直後、他の goroutine から使い始める前に設定してください
	//     (使用中に書き換えるとデータ競合になります)。
	OnEvict func(key K, value V)
}

// New は容量 capacity のキャッシュを作成します。capacity が 1 未満なら panic します。
func New[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity < 1 {
		panic("lru: capacity must be >= 1")
	}
	return &Cache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[K]*list.Element, capacity),
	}
}

// Get はキーに対応する値を返し、その要素を「最近使われた」状態にします。
func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, hit := c.items[key]; hit {
		c.ll.MoveToFront(el)
		return el.Value.(*entry[K, V]).value, true
	}
	return value, false // value はゼロ値
}

// Put は値を追加・更新します。容量を超えた場合は最も古い要素を追い出します。
func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, hit := c.items[key]; hit {
		c.ll.MoveToFront(el)
		el.Value.(*entry[K, V]).value = value
		return
	}
	c.items[key] = c.ll.PushFront(&entry[K, V]{key, value})
	if c.ll.Len() > c.capacity {
		oldest := c.ll.Back()
		e := c.ll.Remove(oldest).(*entry[K, V])
		delete(c.items, e.key)
		if c.OnEvict != nil {
			c.OnEvict(e.key, e.value)
		}
	}
}

// Len は現在の要素数を返します。
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

// All は新しい順にキーと値を列挙するイテレータを返します (Go 1.23+ の range-over-func)。
//
//	for k, v := range c.All() { ... }
//
// 列挙中はロックを保持するため、ループ本体からキャッシュを操作しないでください。
func (c *Cache[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		c.mu.Lock()
		defer c.mu.Unlock()
		for el := c.ll.Front(); el != nil; el = el.Next() {
			e := el.Value.(*entry[K, V])
			if !yield(e.key, e.value) {
				return
			}
		}
	}
}
