# 第6章 パッケージ設計とプロジェクト構成

## 6.1 モジュールとパッケージ

| 単位 | 実体 | 説明 |
| --- | --- | --- |
| **モジュール** | `go.mod` があるディレクトリ以下 | バージョン管理・配布の単位。通常 1 リポジトリ = 1 モジュール |
| **パッケージ** | 1 つのディレクトリ | コンパイル・import の単位 |

```
go.mod
module github.com/you/app   ← モジュールパス
go 1.24                     ← 言語バージョン
require github.com/spf13/cobra v1.10.2
tool golang.org/x/tools/cmd/stringer   ← Go 1.24+: 開発ツールの依存も管理できる
```

`go.sum` は依存のハッシュ値で、**必ずコミット** します。

## 6.2 パッケージ命名

[Effective Go](https://go.dev/doc/effective_go#package-names) / [Go Blog: Package names](https://go.dev/blog/package-names) の要点:

- **短く、小文字、単数形、アンダースコアなし**: `http`, `json`, `user` (✕ `userService`, `utils`)
- 呼び出し側でパッケージ名が付くので **名前の重複を避ける**: `user.User` より `user.Account`、`http.HTTPServer` ではなく `http.Server`
- `util`, `common`, `misc`, `helpers` のような **意味のない名前は避ける**

## 6.3 公式が推奨するレイアウト

Go チームは [Organizing a Go module](https://go.dev/doc/modules/layout) で、規模別の構成を示しています。

### (a) 小さなコマンド — ルートに main.go

```
hello/
├── go.mod
├── main.go
└── auth.go
```

### (b) コマンド + 内部パッケージ

```
app/
├── go.mod
├── main.go
└── internal/
    ├── auth/
    └── hash/
```

### (c) 複数コマンド / サーバー

```
project/
├── go.mod
├── cmd/
│   ├── server/main.go
│   └── migrate/main.go
└── internal/
    ├── book/       ← ドメイン
    └── httpapi/    ← HTTP 層
```

### `internal/` ディレクトリ

`internal/` 以下のパッケージは **親ディレクトリ配下からしか import できない** ことをコンパイラが保証します。外部に公開したくない実装を置く場所で、ライブラリの互換性を守るうえで重要です。

### `pkg/` は必要?

[`golang-standards/project-layout`](https://github.com/golang-standards/project-layout) というリポジトリが有名ですが、**Go 公式ではありません**。Go チームの Russ Cox 自身が「これは標準ではない」と issue でコメントしています。`pkg/` は Kubernetes などの初期の大規模プロジェクトの慣習で、**新規プロジェクトで必須ではありません**。迷ったら公式ガイドの (a)〜(c) に従いましょう。

## 6.4 📦 OSS のプロジェクト構成を調査する

主要 OSS のリポジトリ構成を調べました (2026 年 11 月時点)。

| OSS | エントリポイント | 主な構成 | 特徴 |
| --- | --- | --- | --- |
| **GitHub CLI** (`cli/cli`) | `cmd/gh/main.go` | `internal/` (内部実装), `pkg/cmd/<サブコマンド>/` | `main` は `ghcmd.Main()` を呼ぶだけ。サブコマンドごとにパッケージを分割 |
| **Kubernetes** | `cmd/kubectl/kubectl.go`, `cmd/kube-apiserver/` … | `cmd/`, `pkg/`, `staging/` | 巨大モノレポ。`staging/` 配下を別モジュール (`k8s.io/client-go` 等) として公開 |
| **Prometheus** | `cmd/prometheus/main.go`, `cmd/promtool/main.go` | `tsdb/`, `storage/`, `promql/`, `web/` | 機能ドメインごとにトップレベルパッケージ |
| **Terraform** | `main.go` | `internal/` に **ほぼ全コード** | ライブラリとしての外部利用を意図的に禁止 (2021 年頃に移行) |
| **Gitea** | `main.go` | `cmd/`, `models/`, `modules/`, `routers/`, `services/` | MVC 風のレイヤー分割 |
| **Caddy** | `cmd/caddy/main.go` | ルートが公開パッケージ, `modules/` にプラグイン | ライブラリ + プラグイン機構が前提の設計 |
| **Traefik** | `cmd/traefik/traefik.go` | `pkg/` | `pkg/` 慣習を採用 |
| **Hugo** | `main.go` | `commands/`, `hugolib/`, `resources/` … | ルート main + 機能別パッケージ |
| **fzf** | `main.go` | `src/` | 小規模 CLI。シンプル |
| **lazygit** | `main.go` | `pkg/gui/`, `pkg/commands/` | TUI アプリ |

### 調査から分かる傾向

1. **ルート `main.go` か `cmd/<name>/main.go` の二択**。バイナリが 1 つならルート、複数なら `cmd/`。
2. **`main` はとても薄い**。`main()` は設定読み込みとエントリ関数の呼び出しだけを行い、ロジックはテスト可能なパッケージに置く。
   ```go
   // cli/cli: cmd/gh/main.go (全文)
   func main() {
   	code := ghcmd.Main()
   	os.Exit(int(code))
   }
   ```
3. **`internal/` の活用が増えている**。Terraform のように「全部 internal」にする例もある。外部公開は互換性維持のコストを伴うため。
4. **レイヤー別 (controller/service/repository) より機能別 (user/, order/)** に分けるほうが Go では一般的。標準ライブラリ (`net/http`, `encoding/json`) もそうなっている。

## 6.5 本書のハンズオンの構成

```
handson/
├── todo/            ← (b) ルート main.go + internal/   [第10章]
│   ├── main.go
│   └── internal/{cli,todo}/
├── bookshelf-api/   ← (c) cmd/ + internal/            [第11章]
│   ├── cmd/server/main.go
│   └── internal/{book,httpapi}/
├── linkcheck/       ← (b)                              [第12章]
└── lru/             ← ライブラリ (ルートが公開パッケージ) [第13章]
```

## 6.6 依存の向きを整理する

パッケージ間に **循環 import は許されません**。自然と「ドメイン (内側) ← インフラ・UI (外側)」という向きになるよう設計します。

```
cmd/server  ──▶  internal/httpapi  ──▶  internal/book
   (組み立て)         (HTTP の都合)          (ドメイン + Repository interface)
```

- `book` は HTTP を知らない。`httpapi` は `book.Repository` インターフェースに依存する
- 具体的な実装 (`MemoryRepository` や PostgreSQL 実装) を選んで注入するのは `main` の仕事

DI コンテナは Go では一般的ではなく、**main で手で組み立てる** のが主流です (大規模なら [google/wire](https://github.com/google/wire) や [uber-go/fx](https://github.com/uber-go/fx) という選択肢もあります)。

## 6.7 メジャーバージョンと go.work

- v2 以上の破壊的変更をリリースするときは **モジュールパスに `/v2` を付ける** (Semantic Import Versioning)。例: `github.com/cli/cli/v2`
- 複数モジュールを同時に開発するときは `go work init ./a ./b` でワークスペースを作れる (`go.work` は通常コミットしない)

## ✅ チェックリスト

- [ ] `main` パッケージにロジックを書きすぎていない
- [ ] 外部に公開する必要のないパッケージは `internal/` に置いた
- [ ] パッケージ名に `util` / `common` を使っていない
- [ ] 循環 import が発生しない依存の向きになっている
