package book

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestInput_Validate(t *testing.T) {
	tests := map[string]struct {
		in      Input
		wantErr bool
	}{
		"ok":             {in: Input{Title: "Go", Author: "Gopher"}, wantErr: false},
		"title missing":  {in: Input{Author: "Gopher"}, wantErr: true},
		"author missing": {in: Input{Title: "Go"}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.in.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrValidation) {
				t.Errorf("err should wrap ErrValidation: %v", err)
			}
		})
	}
}

// go test -race で実行すると、データ競合がないことを確認できます。
func TestMemoryRepository_Concurrent(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	for range 100 { // Go 1.22+ の整数 range
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.Create(ctx, Input{Title: "t", Author: "a"}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	books, _ := r.List(ctx)
	if len(books) != 100 {
		t.Fatalf("len = %d, want 100", len(books))
	}
	for i, b := range books {
		if b.ID != int64(i+1) {
			t.Fatalf("books[%d].ID = %d (not sorted / not unique)", i, b.ID)
		}
	}
}
