// Package cli は cobra を使ったコマンド定義をまとめます。
// gh (GitHub CLI) と同様に「コマンドごとに NewCmdXxx 関数」を用意し、
// 入出力や依存 (Store, 現在時刻) を外から注入できるようにしています。
package cli

import (
	"io"
	"os"
	"time"

	"github.com/caffelatte23/golang-handson/handson/todo/internal/todo"
	"github.com/spf13/cobra"
)

// Deps はコマンドが必要とする依存をまとめたものです。テストでは差し替えます。
type Deps struct {
	Out   io.Writer
	Err   io.Writer
	Store todo.Store
	Now   func() time.Time
}

// version は `go build -ldflags "-X .../internal/cli.version=v1.0.0"` で埋め込めます。
var version = "dev"

// NewRootCmd はルートコマンドを組み立てます。
func NewRootCmd(d *Deps) *cobra.Command {
	var file string

	root := &cobra.Command{
		Use:     "todo",
		Short:   "シンプルな TODO 管理 CLI",
		Version: version,
		// エラー時に usage を毎回出さない (gh などと同じ挙動)
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// --file が指定されたときだけ Store を差し替える
			if file != "" {
				d.Store = todo.FileStore{Path: file}
			}
			return nil
		},
	}
	root.SetOut(d.Out)
	root.SetErr(d.Err)
	root.PersistentFlags().StringVarP(&file, "file", "f", "", "保存先 JSON ファイル (default: OS の設定ディレクトリ/todo/tasks.json)")

	root.AddCommand(
		newAddCmd(d),
		newListCmd(d),
		newDoneCmd(d),
		newRemoveCmd(d),
	)
	return root
}

// Execute は main から呼ばれるエントリポイントです。終了コードを返します。
func Execute() int {
	path, err := todo.DefaultPath()
	if err != nil {
		path = "tasks.json"
	}
	d := &Deps{
		Out:   os.Stdout,
		Err:   os.Stderr,
		Store: todo.FileStore{Path: path},
		Now:   time.Now,
	}
	if err := NewRootCmd(d).Execute(); err != nil {
		_, _ = io.WriteString(d.Err, "Error: "+err.Error()+"\n")
		return 1
	}
	return 0
}
