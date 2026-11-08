# 第13章 ハンズオン④ ライブラリ — ジェネリック LRU キャッシュ

最後は「アプリ」ではなく **他の人に import してもらうライブラリ** を作ります。Go 1.18 以降のジェネリクスと 1.23 以降のイテレータを使い、テスト・ドキュメント・ベンチマーク・ファズテストまで揃えた、pkg.go.dev に公開できる品質を目指します。

📁 完成コード: [`handson/lru/`](../handson/lru/)

## 学べること

- 型パラメータを持つ構造体とメソッド (`Cache[K comparable, V any]`)
- `iter.Seq2` を返すイテレータ (range-over-func)
- 標準ライブラリ `container/list` の活用
- ドキュメントコメントと Example テスト (pkg.go.dev に表示される)
- ベンチマーク (`b.Loop`, `-benchmem`)
- ファズテスト (`go test -fuzz`)
- ライブラリの公開とバージョニング

## LRU キャッシュとは

容量が一杯になったとき、**最も長く使われていない (Least Recently Used)** 要素から捨てるキャッシュです。

- **map**: キー → リスト要素 で O(1) 検索
- **双方向連結リスト**: 使われた要素を先頭へ移動 O(1)、末尾を削除 O(1)

を組み合わせると、すべての操作が O(1) になります。

```
最近使った ◀──────────────────────────▶ 古い
[cherry] ⇄ [banana] ⇄ [apple]   ← Put("date") で apple が追い出される
```

## Step 1. モジュールを作る

```bash
mkdir lru && cd lru
go mod init github.com/<your-name>/lru
```

ライブラリなので `main` パッケージは作らず、**ルートディレクトリ自体を公開パッケージ** にします (`import "github.com/<your-name>/lru"`)。

## Step 2. 実装

`lru.go`:

```go
// handson/lru/lru.go
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
```

**ポイント**

- **型パラメータ**: キーは map のキーにするので `comparable` 制約、値は何でもよいので `any`。
- **ゼロ値の返却**: `Get` の名前付き戻り値 `value V` は、ヒットしなかったときに自動的に `V` のゼロ値になります。ジェネリクスでは `var zero V` や名前付き戻り値でゼロ値を作るのが定番です。
- **`container/list`** はジェネリクス以前の標準パッケージなので要素は `any` 型です。そのため `el.Value.(*entry[K, V])` と型アサーションしています。
- **イテレータ `All()`**: `iter.Seq2[K, V]` を返すと、利用者は `for k, v := range c.All()` と書けます。さらに `maps.Collect(c.All())` のように標準ライブラリの関数にもそのまま渡せます。
- **ドキュメントコメント**: パッケージ・公開識別子の直前のコメントは `go doc` や pkg.go.dev にそのまま表示されます。**識別子名で始める** のが慣習です (`// New は...` / `// New creates...`)。
- **panic の使いどころ**: `capacity < 1` はプログラマのミス (バグ) なので panic にしています。`regexp.MustCompile` と同じ考え方です。

## Step 3. 単体テストと Example

`lru_test.go` (抜粋):

```go
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
	// ...
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
```

```bash
go test -race -v ./...
go doc -all .   # ドキュメントの見え方を確認
```

## Step 4. ベンチマーク

```go
func BenchmarkCache_Get(b *testing.B) {
	c := New[int, int](1024)
	for i := range 1024 {
		c.Put(i, i)
	}
	i := 0
	for b.Loop() { // Go 1.24+: セットアップ部分は計測対象外になる
		c.Get(i % 1024)
		i++
	}
}
```

```console
$ go test -bench=. -benchmem -run='^$'
BenchmarkCache_Put-4   	 7927347	       153.0 ns/op	      64 B/op	       2 allocs/op
BenchmarkCache_Get-4   	43785204	        27.52 ns/op	       0 B/op	       0 allocs/op
```

- `ns/op`: 1 操作あたりの時間、`B/op` / `allocs/op`: 1 操作あたりのメモリ確保
- `Get` はアロケーション 0 です。`Put` の 2 allocs は `entry` と `list.Element` の生成によるもの
- 改善を試したら [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) で統計的に比較するのが OSS の作法です (Go 本体の PR でも必須)

## Step 5. ファズテスト

ファズテストは、ランダムに生成した入力で **常に成り立つべき性質 (プロパティ)** を検証します。

```go
func FuzzCache(f *testing.F) {
	f.Add(3, []byte{1, 2, 3, 1, 4, 5}) // シード
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
```

```bash
go test -fuzz=FuzzCache -fuzztime=30s
```

失敗する入力が見つかると `testdata/fuzz/FuzzCache/` に保存され、以後は通常の `go test` で回帰テストとして実行されます。

## Step 6. 公開する

```bash
git tag v0.1.0
git push origin v0.1.0
```

これだけで、誰でも `go get github.com/<your-name>/lru@v0.1.0` で使えるようになります。数分後に [pkg.go.dev](https://pkg.go.dev/) でドキュメントも表示されます (表示されない場合は pkg.go.dev 上で「Request」ボタンを押す)。

**バージョニングのルール** (Semantic Versioning):

| 変更 | バージョン |
| --- | --- |
| バグ修正 | v0.1.0 → v0.1.1 |
| 後方互換な機能追加 | v0.1.0 → v0.2.0 |
| 破壊的変更 (v1 以降) | v1.x → **v2.0.0 + モジュールパスに `/v2`** |

`v0.x` の間は破壊的変更が許容されます。API が固まったら `v1.0.0` を打ちましょう。

## 📦 OSS と比べてみよう

実際の LRU キャッシュライブラリ [`hashicorp/golang-lru`](https://github.com/hashicorp/golang-lru) は、v2 でジェネリクスに対応し、モジュールパスが `github.com/hashicorp/golang-lru/v2` になりました。まさに Step 6 の「破壊的変更 → `/v2`」の実例です。

また、ライブラリの API 設計については以下が参考になります。

- 公開 API は **最小限に** (後から追加はできても、削除は破壊的変更)
- 内部実装は `internal/` か非公開識別子に
- 設定項目が増えそうなら Functional Options (第2章)

## 🚀 発展課題

1. **TTL**: 要素ごとに有効期限を持たせ、期限切れは `Get` で返さないようにしよう (時刻は関数で注入してテスト可能に)。
2. **統計**: ヒット率 (`Hits`, `Misses`) を `atomic.Int64` で数えよう。
3. **シャーディング**: キーのハッシュで N 個の Cache に振り分け、ロック競合を減らそう。ベンチマークで `-cpu=1,4,8` を比較。
4. **singleflight**: 第11章の Web API に組み込み、キャッシュミス時の同時取得を `golang.org/x/sync/singleflight` で 1 回にまとめよう。
