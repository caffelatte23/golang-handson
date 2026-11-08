package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Store はタスクリストの永続化を抽象化したインターフェースです。
// 「使う側」で小さく定義するのが Go 流です。
type Store interface {
	Load() (*List, error)
	Save(*List) error
}

// FileStore は JSON ファイルにタスクを保存します。
type FileStore struct {
	Path string
}

// Load はファイルからリストを読み込みます。ファイルがなければ空のリストを返します。
func (s FileStore) Load() (*List, error) {
	b, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return &List{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", s.Path, err)
	}
	var l List
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("decode %s: %w", s.Path, err)
	}
	return &l, nil
}

// Save はリストをファイルへ書き込みます。
// 一時ファイルに書いてから rename することで、途中で落ちてもファイルが壊れません。
func (s FileStore) Save(l *List) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.Path)
}

// DefaultPath は保存先のデフォルトパスを返します (例: ~/.config/todo/tasks.json)。
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo", "tasks.json"), nil
}
