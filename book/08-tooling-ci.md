# 第8章 ツールチェイン・Lint・CI/CD・リリース

## 8.1 フォーマッタ

| ツール | 内容 |
| --- | --- |
| `gofmt` | 公式フォーマッタ。**設定項目なし**。全員が同じ見た目になる |
| `goimports` | gofmt + import の自動追加・削除・グルーピング |
| `gofumpt` | gofmt より厳しい上位互換 (一部 OSS が採用) |

> *Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.* — Go Proverbs

## 8.2 静的解析

| ツール | 内容 |
| --- | --- |
| `go vet` | 公式。Printf の引数ミス、到達しないコード、構造体タグの誤りなど |
| [`staticcheck`](https://staticcheck.dev/) | 最も定評のある単体 linter。バグ・非推奨 API・簡略化提案 |
| [`golangci-lint`](https://golangci-lint.run/) | 100 以上の linter を束ねて並列実行するランナー。**事実上の標準** |
| [`govulncheck`](https://go.dev/doc/security/vuln/) | 公式の脆弱性スキャナ。**実際に呼んでいる関数** に脆弱性があるかまで解析 |

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### 📦 OSS の golangci-lint 設定を調査する

調査した 15 リポジトリのうち、**cli/cli, prometheus, traefik, gitea, caddy, glow, minio, lazygit, containerd** の 9 つがリポジトリ直下に `.golangci.yml` を持っていました。

GitHub CLI の設定 (抜粋) を見てみましょう。v2 形式で、`default: none` から必要なものを明示的に有効化しています。

```yaml
version: "2"
linters:
  default: none
  enable:
    - bodyclose     # HTTP レスポンスボディの Close 漏れ
    - copyloopvar   # Go 1.22 以降不要なループ変数コピー
    - gocritic
    - govet
    - ineffassign
    - nilerr        # err != nil なのに nil を返す
    - unused
    # ...
    # To enable later due to too many issues, and confirm we need them:
    # - gosec
    # - staticcheck
    # - errcheck
```

**学べること**:

- 既存プロジェクトに後から入れる場合、**いきなり全部有効にしない**。ノイズの少ない linter から段階的に増やす (cli/cli は「後で有効化する」リストをコメントで残している)
- テストファイルは一部 linter を除外する (`exclusions.rules` で `path: _test\.go$`)

本リポジトリの [`.golangci.yml`](../.golangci.yml) は、新規プロジェクト向けに `default: standard` (errcheck, govet, ineffassign, staticcheck, unused) をベースに、`bodyclose`, `errorlint`, `gocritic`, `revive` などを追加した構成です。

```bash
golangci-lint run ./...
```

## 8.3 タスクランナー

| 方式 | 採用 OSS (調査対象のうち) | 特徴 |
| --- | --- | --- |
| **Makefile** | cli/cli, Kubernetes, Prometheus, Terraform, Traefik, Gitea, fzf, etcd, k6, MinIO, lazygit, containerd | 最も一般的。どの環境にもある |
| [Mage](https://magefile.org/) (`magefile.go`) | Hugo | タスクを Go で書ける |
| [Task](https://taskfile.dev/) (`Taskfile.yml`) | — (新興プロジェクトで増加中) | YAML で書けるクロスプラットフォームな make 代替 |

Go のビルド自体は `go build` だけで完結するため、Makefile は「`lint` / `test` / `generate` などのショートカット集」として使われることがほとんどです。本リポジトリの [`Makefile`](../Makefile) も参考にしてください。

### 開発ツールのバージョン管理 (Go 1.24+)

以前は `tools.go` に blank import する裏技が使われていましたが、Go 1.24 で `tool` ディレクティブが入りました。

```bash
go get -tool golang.org/x/tools/cmd/stringer   # go.mod に tool 行が追加される
go tool stringer -type=Status                  # 固定バージョンで実行
```

## 8.4 GitHub Actions で CI

本リポジトリの [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) の要点:

```yaml
steps:
  - uses: actions/checkout@v4
  - uses: actions/setup-go@v5
    with:
      go-version-file: go.mod   # go.mod の go 行に合わせる (バージョンの二重管理を避ける)
  - run: go vet ./...
  - run: go test -race -coverprofile=coverage.out ./...
  - uses: golangci/golangci-lint-action@v8
    with:
      version: v2.5
```

よく追加されるジョブ:

- **OS マトリクス** (`ubuntu`, `macos`, `windows`): CLI ツールでは必須級 (gh, fzf, lazygit)
- **govulncheck**: `golang/govulncheck-action`
- **`go mod tidy` の差分チェック**: `go mod tidy && git diff --exit-code`
- **生成コードの差分チェック**: `go generate ./... && git diff --exit-code`

### 依存の自動更新

| ツール | 採用 OSS (調査対象のうち) |
| --- | --- |
| Dependabot (`.github/dependabot.yml`) | cli/cli, Hugo, Terraform, Caddy, fzf, Glow, etcd, lazygit, containerd |
| Renovate (`renovate.json`) | Prometheus, k6 |

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule: { interval: weekly }
  - package-ecosystem: github-actions
    directory: /
    schedule: { interval: weekly }
```

## 8.5 バージョン情報の埋め込み

ビルド時に `-ldflags -X` で変数を上書きするのが定番です。

```go
package build
var Version = "dev"
```

```bash
go build -ldflags "-X github.com/you/app/internal/build.Version=v1.2.3" ./cmd/app
```

GitHub CLI の `.goreleaser.yml` でも同じ手法が使われています:

```yaml
ldflags:
  - -s -w -X github.com/cli/cli/v2/internal/build.Version={{.Version}} -X github.com/cli/cli/v2/internal/build.Date={{time "2006-01-02"}}
```

> 💡 `-s -w` はデバッグ情報を削ってバイナリを小さくします。また Go 1.18 以降は `runtime/debug.ReadBuildInfo()` で VCS のコミットハッシュなどを取得できるので、ldflags なしでもある程度の情報は得られます。

## 8.6 リリース: GoReleaser

[GoReleaser](https://goreleaser.com/) は、タグを打つだけで **複数 OS/アーキテクチャのバイナリ、アーカイブ、チェックサム、GitHub Release、Homebrew tap、Docker イメージ、deb/rpm** までまとめて作ってくれるツールです。

調査対象では **cli/cli, Caddy, fzf, Glow, lazygit** が `.goreleaser.yml` を持っていました。**配布される CLI ツールではほぼ標準** と言ってよいでしょう。

→ 具体的な設定は第14章で扱います。

## ✅ チェックリスト

- [ ] 保存時 `goimports` (または gofmt) が走る
- [ ] `golangci-lint run` が CI で動いている
- [ ] `go test -race` が CI で動いている
- [ ] Dependabot / Renovate で依存が自動更新される
