# 第14章 ハンズオン① CLI ツール — TODO 管理 (cobra)

Go は **CLI ツールの開発言語として最も人気のある言語の一つ** です。単一バイナリで配布でき、起動が速く、クロスコンパイルが簡単だからです。この章では GitHub CLI や kubectl と同じ **cobra** を使って、サブコマンド型の TODO 管理ツールを作ります。

📁 完成コード: [`handson/todo/`](../handson/todo/)

## 学べること

- cobra によるサブコマンド・フラグ・引数バリデーション
- ドメインロジックと CLI 層の分離 (`internal/todo` と `internal/cli`)
- `io.Writer` とインターフェースによる **依存注入** とテスト
- JSON ファイルへの安全な永続化 (一時ファイル + rename)
- `text/tabwriter` による表形式出力、`--json` 出力
- `-ldflags` によるバージョン埋め込み

## 完成イメージ

```console
$ todo add 牛乳を買う
added #1: 牛乳を買う
$ todo add 本を読む
added #2: 本を読む
$ todo done 1
completed #1
$ todo ls --all
ID  STATUS  TITLE
1   [x]     牛乳を買う
2   [ ]     本を読む
$ todo rm 9
Error: id=9: task not found
$ echo $?
1
```

## Step 1. プロジェクトを作る

```bash
mkdir todo && cd todo
go mod init github.com/<your-name>/todo
mkdir -p internal/todo internal/cli
```

最終的な構成:

```
todo/
├── go.mod
├── main.go                 ← エントリポイント (薄く保つ)
└── internal/
    ├── todo/               ← ドメイン: CLI を知らない
    │   ├── todo.go
    │   ├── store.go
    │   └── todo_test.go
    └── cli/                ← cobra のコマンド定義
        ├── root.go
        ├── commands.go
        └── cli_test.go
```

## Step 2. ドメインモデルを書く

まずは CLI のことは忘れて、**タスクの追加・完了・削除** というロジックだけを書きます。

`internal/todo/todo.go`:

```go
// handson/todo/internal/todo/todo.go
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
```

**ポイント**

- `List` は **ゼロ値で使える** (`var l todo.List` で即 `Add` できる)。`NextID == 0` のとき 1 から始める工夫をしています。
- 現在時刻を引数 `now time.Time` で受け取ることで、テストで時刻を固定できます。`time.Now()` を内部で直接呼ぶとテストしづらくなります。
- `ErrNotFound` は **センチネルエラー**、`fmt.Errorf("...: %w", ...)` でラップして ID 情報を付け足しています (第7章)。

## Step 3. テストを書く

`internal/todo/todo_test.go` (抜粋):

```go
func TestList_Add(t *testing.T) {
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
			// ...
		})
	}
}
```

```bash
go test ./internal/todo/ -v
```

## Step 4. 永続化 (JSON ファイル)

`internal/todo/store.go`:

```go
// handson/todo/internal/todo/store.go
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
```

**ポイント**

- `Store` インターフェースは **2 メソッドだけ**。CLI 層はこのインターフェースにだけ依存するので、テストではメモリ上のフェイクに差し替えられます。
- 書き込みは「一時ファイルに書く → `os.Rename`」。rename は同一ファイルシステム上でアトミックなので、書き込み途中でプロセスが落ちても元のファイルが壊れません。
- `os.UserConfigDir()` は OS ごとの設定ディレクトリ (Linux: `~/.config`, macOS: `~/Library/Application Support`, Windows: `%AppData%`) を返します。
- ファイルが無いことは `errors.Is(err, fs.ErrNotExist)` で判定します。

## Step 5. cobra を導入してルートコマンドを作る

```bash
go get github.com/spf13/cobra@latest
```

`internal/cli/root.go`:

```go
// handson/todo/internal/cli/root.go
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
```

**ポイント**

- `Deps` 構造体に **出力先・Store・時刻関数** をまとめ、各コマンドに渡します。これは GitHub CLI の `cmdutil.Factory` / `IOStreams` と同じ考え方です。
- `SilenceUsage: true` を付けないと、エラーのたびに長い usage が表示されてしまいます。
- `PersistentFlags` は全サブコマンドで使えるフラグ、`Flags` はそのコマンドだけのフラグです。
- `PersistentPreRunE` はサブコマンド実行前に呼ばれるフックです。

## Step 6. サブコマンドを実装する

`internal/cli/commands.go`:

```go
// handson/todo/internal/cli/commands.go
package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/caffelatte23/golang-handson/handson/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newAddCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:     "add <title>...",
		Short:   "タスクを追加する",
		Example: `  todo add 牛乳を買う`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return update(d, func(l *todo.List) error {
				t, err := l.Add(strings.Join(args, " "), d.Now())
				if err != nil {
					return err
				}
				fmt.Fprintf(d.Out, "added #%d: %s\n", t.ID, t.Title)
				return nil
			})
		},
	}
}

func newListCmd(d *Deps) *cobra.Command {
	var (
		all    bool
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "タスクを一覧表示する",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := d.Store.Load()
			if err != nil {
				return err
			}
			tasks := l.Filter(func(t todo.Task) bool { return all || !t.Done })

			if asJSON {
				enc := json.NewEncoder(d.Out)
				enc.SetIndent("", "  ")
				return enc.Encode(tasks)
			}
			if len(tasks) == 0 {
				fmt.Fprintln(d.Out, "タスクはありません 🎉")
				return nil
			}
			// tabwriter で列をそろえる (kubectl get などと同じ手法)
			w := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tTITLE")
			for _, t := range tasks {
				status := "[ ]"
				if t.Done {
					status = "[x]"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\n", t.ID, status, t.Title)
			}
			return w.Flush()
		},
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "完了済みのタスクも表示する")
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON で出力する")
	return cmd
}

func newDoneCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "タスクを完了にする",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			return update(d, func(l *todo.List) error {
				if err := l.Complete(id, d.Now()); err != nil {
					return err
				}
				fmt.Fprintf(d.Out, "completed #%d\n", id)
				return nil
			})
		},
	}
}

func newRemoveCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <id>",
		Aliases: []string{"rm"},
		Short:   "タスクを削除する",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			return update(d, func(l *todo.List) error {
				if err := l.Remove(id); err != nil {
					return err
				}
				fmt.Fprintf(d.Out, "removed #%d\n", id)
				return nil
			})
		},
	}
}

// update は「読み込み → 変更 → 保存」の共通処理です。
func update(d *Deps, fn func(*todo.List) error) error {
	l, err := d.Store.Load()
	if err != nil {
		return err
	}
	if err := fn(l); err != nil {
		return err
	}
	return d.Store.Save(l)
}

func parseID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid id %q: must be a number", s)
	}
	return id, nil
}
```

**ポイント**

- `Args: cobra.ExactArgs(1)` / `MinimumNArgs(1)` / `NoArgs` で引数の数を宣言的に検証できます。
- `RunE` (エラーを返す版) を使い、エラー処理は最上位の `Execute()` に集約します。
- 「読み込み → 変更 → 保存」の共通部分を `update` ヘルパーに切り出し、クロージャで差分だけ渡しています。
- `--json` 出力を用意しておくと、`jq` などと組み合わせられて便利です (gh の `--json` フラグと同じ発想)。

## Step 7. main.go

```go
// handson/todo/main.go
// todo は cobra を使った TODO 管理 CLI のサンプルです。
package main

import (
	"os"

	"github.com/caffelatte23/golang-handson/handson/todo/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
```

`os.Exit` は `defer` を実行せずに終了するため、`main` には `os.Exit` 以外のロジックを置かないのが定石です。

## Step 8. CLI 全体をテストする

`internal/cli/cli_test.go`:

```go
// handson/todo/internal/cli/cli_test.go
package cli

import (
	"bytes"
	"errors"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/caffelatte23/golang-handson/handson/todo/internal/todo"
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
```

**ポイント**

- `cmd.SetArgs(...)` で引数を与え、`bytes.Buffer` に出力を集めて検証します。**プロセスを起動せずに** CLI 全体をテストできます。
- `memStore.Load` で `slices.Clone` しているのは、スライスが内部配列を共有するため。コピーしないと「保存していないのに変更が反映される」バグを見逃します。

```bash
go test ./... -race
```

## Step 9. ビルドと配布

```bash
# ビルド
go build -o todo .
./todo --help

# バージョンを埋め込む
go build -ldflags "-X github.com/<your-name>/todo/internal/cli.version=v0.1.0" -o todo .
./todo --version   # todo version v0.1.0

# $GOPATH/bin にインストール
go install .

# シェル補完 (cobra が自動生成)
source <(todo completion bash)
```

公開リポジトリにプッシュすれば、誰でも次の 1 行でインストールできます。

```bash
go install github.com/<your-name>/todo@latest
```

> 💡 本リポジトリの `handson/todo` のように、**リポジトリのサブディレクトリにあるモジュール** をバージョン付きで公開するときは、タグに **ディレクトリ名を前に付けます** (`handson/todo/v0.1.0`)。リポジトリ直下のモジュールなら、単に `v0.1.0` です。

## 📦 OSS と比べてみよう — GitHub CLI のコマンド構造

GitHub CLI (`cli/cli`) は、cobra の使い方のお手本です。

```
pkg/cmd/
├── root/root.go         ← NewCmdRoot(f *cmdutil.Factory, ...)
├── issue/
│   ├── issue.go         ← NewCmdIssue: サブコマンドを束ねる
│   ├── create/create.go ← NewCmdCreate(f, runF): 1 コマンド = 1 パッケージ
│   └── list/list.go
└── pr/...
```

各コマンドは次の形で書かれています (簡略化):

```go
type CreateOptions struct {
	IO         *iostreams.IOStreams
	HttpClient func() (*http.Client, error)
	Title      string
	// ...
}

func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{IO: f.IOStreams, HttpClient: f.HttpClient}
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts) // テスト時: フラグのパース結果だけを検証できる
			}
			return createRun(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Title, "title", "t", "", "...")
	return cmd
}
```

- **Options 構造体** にフラグと依存をまとめる
- **`runF` を差し替え可能** にすることで「フラグのパース」と「実行ロジック」を別々にテストできる
- `Factory` が HTTP クライアントや設定を遅延生成する

本章の TODO CLI はこれを簡略化した形です。規模が大きくなったら、この構造に寄せていくと良いでしょう。

## 🚀 発展課題

1. `todo edit <id> <title>` コマンドを追加しよう。
2. 期限 (`--due 2026-12-31`) と優先度を持てるようにし、`list --sort due` で並べ替えよう (`slices.SortFunc`)。
3. 設定ファイル (`~/.config/todo/config.yaml`) を [viper](https://github.com/spf13/viper) で読み込もう。
4. [bubbletea](https://github.com/charmbracelet/bubbletea) で対話的な TUI モード (`todo ui`) を作ろう。
5. 保存先を SQLite (`modernc.org/sqlite`) に差し替えた `SQLiteStore` を実装しよう。`Store` インターフェースのおかげで CLI 層は変更不要なはずです。
