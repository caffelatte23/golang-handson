// Package todo は TODO リストのドメインロジックを提供します。
// CLI や保存形式には依存しないため、単体でテストできます。
package todo

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound は指定 ID のタスクが存在しないことを表すセンチネルエラーです。
// 呼び出し側は errors.Is(err, todo.ErrNotFound) で判定します。
var ErrNotFound = errors.New("task not found")

// Task は 1 件のタスクです。
type Task struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// List はタスクの集合です。ゼロ値のまま使えます。
type List struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

// Add はタスクを追加し、追加されたタスクを返します。
func (l *List) Add(title string, now time.Time) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, errors.New("title must not be empty")
	}
	if l.NextID == 0 {
		l.NextID = 1
	}
	t := Task{ID: l.NextID, Title: title, CreatedAt: now}
	l.Tasks = append(l.Tasks, t)
	l.NextID++
	return t, nil
}

// Complete は指定 ID のタスクを完了にします。
func (l *List) Complete(id int, now time.Time) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	l.Tasks[i].Done = true
	l.Tasks[i].DoneAt = &now
	return nil
}

// Remove は指定 ID のタスクを削除します。
func (l *List) Remove(id int) error {
	i, err := l.index(id)
	if err != nil {
		return err
	}
	l.Tasks = append(l.Tasks[:i], l.Tasks[i+1:]...)
	return nil
}

// Filter は条件に合うタスクだけを返します。
func (l *List) Filter(keep func(Task) bool) []Task {
	var out []Task
	for _, t := range l.Tasks {
		if keep(t) {
			out = append(out, t)
		}
	}
	return out
}

func (l *List) index(id int) (int, error) {
	for i, t := range l.Tasks {
		if t.ID == id {
			return i, nil
		}
	}
	// %w でラップすると、メッセージを補足しつつ errors.Is で判定できる。
	return -1, fmt.Errorf("id=%d: %w", id, ErrNotFound)
}
