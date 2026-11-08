// Package book は書籍のドメインモデルとリポジトリを定義します。
package book

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("book not found")
	ErrValidation = errors.New("validation error")
)

// Book は書籍を表すエンティティです。
type Book struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input は作成・更新時の入力です。
type Input struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Read   bool   `json:"read"`
}

// Validate は入力を検証します。エラーは ErrValidation をラップします。
func (in *Input) Validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Author = strings.TrimSpace(in.Author)
	var problems []string
	if in.Title == "" {
		problems = append(problems, "title is required")
	}
	if len([]rune(in.Title)) > 200 {
		problems = append(problems, "title must be <= 200 characters")
	}
	if in.Author == "" {
		problems = append(problems, "author is required")
	}
	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}

// ValidationError は複数の検証エラーを保持するカスタムエラー型です。
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "validation error: " + strings.Join(e.Problems, ", ")
}

// Is を実装すると errors.Is(err, ErrValidation) が true になります。
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

// Repository は永続化層の抽象です。
// 実装はインメモリ (memory.go) ですが、PostgreSQL などに差し替えられます。
type Repository interface {
	List(ctx context.Context) ([]Book, error)
	Get(ctx context.Context, id int64) (Book, error)
	Create(ctx context.Context, in Input) (Book, error)
	Update(ctx context.Context, id int64, in Input) (Book, error)
	Delete(ctx context.Context, id int64) error
}
