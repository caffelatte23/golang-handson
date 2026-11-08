package cli

import (
	"bytes"
	"errors"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/caffelatte23/golang-hanson/handson/todo/internal/todo"
)

// memStore はテスト用のインメモリ Store です。
// インターフェースで依存を受け取っているので、簡単に差し替えられます。
type memStore struct{ l todo.List }

func (m *memStore) Load() (*todo.List, error) {
	c := m.l
	c.Tasks = slices.Clone(m.l.Tasks) // スライスは参照を共有するのでコピーする
	return &c, nil
}
func (m *memStore) Save(l *todo.List) error { m.l = *l; return nil }

func run(t *testing.T, s todo.Store, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	d := &Deps{
		Out:   &out,
		Err:   &out,
		Store: s,
		Now:   func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	}
	cmd := NewRootCmd(d)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestCLI_Flow(t *testing.T) {
	s := &memStore{}

	out, err := run(t, s, "add", "牛乳を買う")
	if err != nil || !strings.Contains(out, "added #1: 牛乳を買う") {
		t.Fatalf("add: out=%q err=%v", out, err)
	}
	if _, err := run(t, s, "add", "本を読む"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, s, "done", "1"); err != nil {
		t.Fatal(err)
	}

	out, _ = run(t, s, "list")
	if strings.Contains(out, "牛乳") || !strings.Contains(out, "本を読む") {
		t.Errorf("list should hide done tasks:\n%s", out)
	}
	out, _ = run(t, s, "list", "--all")
	if !regexp.MustCompile(`1\s+\[x\]\s+牛乳を買う`).MatchString(out) {
		t.Errorf("list --all:\n%s", out)
	}
}

func TestCLI_Errors(t *testing.T) {
	s := &memStore{}
	if _, err := run(t, s, "done", "abc"); err == nil {
		t.Error("want error for non-numeric id")
	}
	if _, err := run(t, s, "rm", "42"); !errors.Is(err, todo.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if _, err := run(t, s, "add"); err == nil {
		t.Error("want error for missing args")
	}
}
