package todo

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

var now = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func TestList_Add(t *testing.T) {
	// テーブル駆動テスト: Go の OSS で最もよく見かける書き方。
	tests := []struct {
		name    string
		title   string
		want    string
		wantErr bool
	}{
		{name: "通常", title: "牛乳を買う", want: "牛乳を買う"},
		{name: "前後の空白は除去", title: "  掃除  ", want: "掃除"},
		{name: "空文字はエラー", title: "   ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l List
			got, err := l.Add(tt.title, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Title != tt.want {
				t.Errorf("Title = %q, want %q", got.Title, tt.want)
			}
			if got.ID != 1 {
				t.Errorf("ID = %d, want 1", got.ID)
			}
		})
	}
}

func TestList_CompleteAndRemove(t *testing.T) {
	var l List
	a, _ := l.Add("a", now)
	b, _ := l.Add("b", now)

	if err := l.Complete(a.ID, now); err != nil {
		t.Fatal(err)
	}
	if !l.Tasks[0].Done {
		t.Error("task a should be done")
	}
	if err := l.Remove(b.ID); err != nil {
		t.Fatal(err)
	}
	if len(l.Tasks) != 1 {
		t.Fatalf("len = %d, want 1", len(l.Tasks))
	}
	if err := l.Remove(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestFileStore_RoundTrip(t *testing.T) {
	// t.TempDir() はテスト終了時に自動で削除される一時ディレクトリ。
	s := FileStore{Path: filepath.Join(t.TempDir(), "tasks.json")}

	l, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Tasks) != 0 {
		t.Fatalf("new store should be empty")
	}
	if _, err := l.Add("永続化する", now); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(l); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 1 || got.Tasks[0].Title != "永続化する" || got.NextID != 2 {
		t.Errorf("loaded = %+v", got)
	}
}
