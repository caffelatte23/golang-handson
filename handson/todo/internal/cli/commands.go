package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/caffelatte23/golang-hanson/handson/todo/internal/todo"
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
