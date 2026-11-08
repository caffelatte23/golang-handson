# 第7章 テスト

Go はテストフレームワークを **標準で内蔵** しています。`xxx_test.go` に `TestXxx(t *testing.T)` を書いて `go test` するだけです。

## 7.1 基本

```go
// calc.go
package calc
func Add(a, b int) int { return a + b }

// calc_test.go
package calc

import "testing"

func TestAdd(t *testing.T) {
	got := Add(1, 2)
	if got != 3 {
		t.Errorf("Add(1, 2) = %d, want 3", got)
	}
}
```

```bash
go test ./...                 # 全パッケージ
go test -run TestAdd -v ./... # 名前で絞り込み、詳細表示
go test -race ./...           # データ競合検出
go test -cover ./...          # カバレッジ
go test -coverprofile=c.out ./... && go tool cover -html=c.out
```

- メッセージは **`got X, want Y`** の形式が慣習 ([Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#useful-test-failures))
- `t.Errorf` は失敗を記録して続行、`t.Fatalf` は即終了

## 7.2 テーブル駆動テスト

**Go で最も重要なテストパターン** です。入力と期待値の表を作り、ループで回します。

```go
func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int64
		wantErr bool
	}{
		{name: "bytes", in: "100", want: 100},
		{name: "kilobytes", in: "2KB", want: 2048},
		{name: "invalid", in: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // サブテスト
			got, err := ParseSize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
```

`go test -run 'TestParseSize/invalid'` のように特定のケースだけ実行できます。

## 7.3 便利な機能

| 機能 | 用途 |
| --- | --- |
| `t.Helper()` | ヘルパー関数内で呼ぶと、失敗行が呼び出し元で報告される |
| `t.Cleanup(fn)` | テスト終了時の後片付け |
| `t.TempDir()` | 自動削除される一時ディレクトリ |
| `t.Setenv(k, v)` | テスト中だけ環境変数を設定 |
| `t.Parallel()` | 並列実行 |
| `t.Context()` (1.24+) | テスト終了時にキャンセルされる context |
| `testing/fstest` | ファイルシステムのモック |
| `net/http/httptest` | HTTP サーバー/レスポンスのテスト |
| `testing/synctest` (1.25+) | 時間に依存する並行処理を決定的にテスト |

## 7.4 HTTP のテスト

```go
func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}
```

実際のサーバーを立てる `httptest.NewServer(handler)` もあり、第11章・第12章のハンズオンで使います。

## 7.5 モックはインターフェースで

Go では「モックライブラリありき」ではなく、**小さなインターフェース + 手書きのフェイク** が基本です。

```go
type memStore struct{ data todo.List }
func (m *memStore) Load() (*todo.List, error) { c := m.data; return &c, nil }
func (m *memStore) Save(l *todo.List) error   { m.data = *l; return nil }
```

インターフェースが大きい場合は [`go.uber.org/mock`](https://github.com/uber-go/mock) (gomock の後継) の `mockgen` で生成する方法もあります。

## 7.6 Example テスト・ベンチマーク・ファズ

```go
// Example: ドキュメントに載り、Output と一致するか検証される
func ExampleAdd() {
	fmt.Println(Add(1, 2))
	// Output: 3
}

// ベンチマーク: go test -bench=. -benchmem
func BenchmarkAdd(b *testing.B) {
	for b.Loop() { // Go 1.24+ (従来は for i := 0; i < b.N; i++)
		Add(1, 2)
	}
}

// ファズテスト (1.18+): go test -fuzz=FuzzReverse
func FuzzReverse(f *testing.F) {
	f.Add("hello")
	f.Fuzz(func(t *testing.T, s string) {
		if Reverse(Reverse(s)) != s {
			t.Errorf("double reverse mismatch: %q", s)
		}
	})
}
```

→ 第13章で 3 つすべてを実装します。

## 7.7 📦 OSS のテスト事情

2026 年 11 月時点の `go.mod` (直接依存) を調査した結果です。

| ライブラリ | 用途 | 採用 OSS (調査対象のうち) |
| --- | --- | --- |
| 標準 `testing` のみ | — | fzf (外部テスト依存なし) |
| [`stretchr/testify`](https://github.com/stretchr/testify) | `assert.Equal`, `require.NoError` などのアサーション | GitHub CLI, Kubernetes, Prometheus, etcd, Caddy, containerd, Gitea, k6, Traefik, lazygit |
| [`google/go-cmp`](https://github.com/google/go-cmp) | 構造体の深い比較と差分表示 (`cmp.Diff`) | Kubernetes, Terraform, Hugo, Prometheus, GitHub CLI, containerd |
| [`onsi/ginkgo`](https://github.com/onsi/ginkgo) + gomega | BDD スタイル | Kubernetes (e2e テスト) |
| [`go.uber.org/mock`](https://github.com/uber-go/mock) | モック生成 | Terraform |

**ポイント**:

- アサーションライブラリは **testify が圧倒的多数派**。ただし Go チーム自身や Google スタイルガイドは「標準 `testing` + `cmp.Diff`」を推奨しており、どちらも正解です。
- 迷ったら、**まず標準だけで書く → 比較が辛くなったら go-cmp → チームの好みで testify** の順で導入すると良いでしょう。本書のハンズオンは標準 `testing` のみで書いています。

```go
// go-cmp の例: 差分が見やすい
if diff := cmp.Diff(want, got); diff != "" {
	t.Errorf("mismatch (-want +got):\n%s", diff)
}

// testify の例
require.NoError(t, err)
assert.Equal(t, want, got)
```

### ゴールデンファイルテスト

出力が大きいコマンド (CLI の help やコード生成結果) は、期待値を `testdata/*.golden` に保存し、`-update` フラグで更新する手法がよく使われます (Go 本体の `gofmt` テスト、Hugo、Terraform など)。`testdata/` ディレクトリは go ツールに無視される特別な名前です。

## ✍️ 練習問題

1. 第2章で作った `WordCount` にテーブル駆動テストを書こう。
2. `go test -cover` でカバレッジを 80% 以上にしてみよう。
