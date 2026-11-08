package book

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"time"
)

// MemoryRepository はスレッドセーフなインメモリ実装です。
// net/http のハンドラは並行に呼ばれるため、mutex で保護します。
type MemoryRepository struct {
	mu     sync.RWMutex
	books  map[int64]Book
	nextID int64
	now    func() time.Time
}

// コンパイル時にインターフェースを満たしているか確認するイディオム。
var _ Repository = (*MemoryRepository)(nil)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{books: make(map[int64]Book), nextID: 1, now: time.Now}
}

func (r *MemoryRepository) List(ctx context.Context) ([]Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Book, 0, len(r.books))
	for _, b := range r.books {
		out = append(out, b)
	}
	// map の反復順は不定なので ID でソートして返す
	slices.SortFunc(out, func(a, b Book) int { return cmp.Compare(a.ID, b.ID) })
	return out, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id int64) (Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	return b, nil
}

func (r *MemoryRepository) Create(ctx context.Context, in Input) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	b := Book{ID: r.nextID, Title: in.Title, Author: in.Author, Read: in.Read, CreatedAt: now, UpdatedAt: now}
	r.books[b.ID] = b
	r.nextID++
	return b, nil
}

func (r *MemoryRepository) Update(ctx context.Context, id int64, in Input) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	b.Title, b.Author, b.Read, b.UpdatedAt = in.Title, in.Author, in.Read, r.now()
	r.books[id] = b
	return b, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}
