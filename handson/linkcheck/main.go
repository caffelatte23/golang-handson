// linkcheck は URL のリストを並行にチェックする CLI です。
//
//	linkcheck https://go.dev https://example.com/404
//	cat urls.txt | linkcheck -c 16 -timeout 3s
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/caffelatte23/golang-handson/handson/linkcheck/internal/checker"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// 標準ライブラリ flag だけでも小さな CLI は十分書ける
	fs := flag.NewFlagSet("linkcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	concurrency := fs.Int("c", 8, "同時実行数")
	timeout := fs.Duration("timeout", 10*time.Second, "全体のタイムアウト")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	urls := fs.Args()
	if len(urls) == 0 { // 引数がなければ標準入力から 1 行 1 URL で読む
		sc := bufio.NewScanner(stdin)
		for sc.Scan() {
			if line := strings.TrimSpace(sc.Text()); line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintln(stderr, "read stdin:", err)
			return 2
		}
	}
	if len(urls) == 0 {
		fs.Usage()
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, *timeout)
	defer cancel()

	c := &checker.Checker{
		Client:      &http.Client{Timeout: 5 * time.Second},
		Concurrency: *concurrency,
	}
	results, err := c.Check(ctx, urls)

	broken := 0
	for _, r := range results {
		switch {
		case r.URL == "":
			continue // 中断されたため実行されなかった URL
		case r.Err != nil:
			broken++
			fmt.Fprintf(stdout, "✗ ERR  %-50s %v\n", r.URL, r.Err)
		case !r.OK():
			broken++
			fmt.Fprintf(stdout, "✗ %d  %-50s %v\n", r.Status, r.URL, r.Duration.Round(time.Millisecond))
		default:
			fmt.Fprintf(stdout, "✓ %d  %-50s %v\n", r.Status, r.URL, r.Duration.Round(time.Millisecond))
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	fmt.Fprintf(stdout, "\n%d checked, %d broken\n", len(results), broken)
	if broken > 0 {
		return 1
	}
	return 0
}
