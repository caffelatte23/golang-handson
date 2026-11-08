# 第13章 エコシステムと定番ライブラリ

Go は「まず標準ライブラリで書く」文化が強い言語ですが、分野ごとに定番の外部ライブラリも存在します。この章では、主要 OSS の `go.mod` を調査した結果をもとに、用途別の選択肢を整理します。

## 調査方法

2026 年 11 月時点で、以下 15 リポジトリのデフォルトブランチの `go.mod` から **直接依存 (`// indirect` でないもの)** を抽出しました。

> cli/cli, kubernetes/kubernetes, prometheus/prometheus, hashicorp/terraform, gohugoio/hugo, etcd-io/etcd, containerd/containerd, caddyserver/caddy, traefik/traefik, go-gitea/gitea, grafana/k6, minio/minio, junegunn/fzf, jesseduffield/lazygit, charmbracelet/glow

あわせて、各 `go.mod` の `go` ディレクティブは 1.23〜1.27 に分布しており、**多くの大型 OSS が最新〜1 つ前のリリースに追従している** ことも分かりました。

## 13.1 CLI フレームワーク

| ライブラリ | 採用 OSS (調査対象のうち) | 特徴 |
| --- | --- | --- |
| [`spf13/cobra`](https://github.com/spf13/cobra) | **GitHub CLI, kubectl/Kubernetes, Hugo, etcd, Caddy, k6, Glow** | サブコマンド型 CLI の事実上の標準。補完スクリプトやドキュメント生成も可能 |
| [`urfave/cli`](https://github.com/urfave/cli) (v3) | Gitea, containerd | 宣言的に書ける。cobra に次ぐ人気 |
| [`alecthomas/kingpin`](https://github.com/alecthomas/kingpin) | Prometheus | 古参。後継は [`alecthomas/kong`](https://github.com/alecthomas/kong) (構造体タグで定義) |
| 標準 `flag` | (小規模な CLI 全般) | サブコマンドが不要なら十分 |
| その他 | fzf (引数を自前で解析)、lazygit ([`integrii/flaggy`](https://github.com/integrii/flaggy)) | 依存を減らしたい・独自の要件がある場合 |
| [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) + lipgloss | GitHub CLI, Glow | リッチな TUI (Elm アーキテクチャ) |

→ **第14章で cobra、第16章で標準 `flag`** を使った CLI を作ります。

## 13.2 設定管理

| ライブラリ | 特徴 |
| --- | --- |
| 標準 `os.Getenv` / `flag` | 小規模ならこれで十分 (第15章の Web API はこれ) |
| [`spf13/viper`](https://github.com/spf13/viper) | 設定ファイル・環境変数・フラグを統合。cobra と相性が良い (Glow が採用) |
| [`knadh/koanf`](https://github.com/knadh/koanf) | viper より軽量・モジュラー |
| [`caarlos0/env`](https://github.com/caarlos0/env), [`kelseyhightower/envconfig`](https://github.com/kelseyhightower/envconfig) | 環境変数を構造体タグでマッピング (12-Factor App 向け) |

## 13.3 HTTP サーバー / ルーター

Go 1.22 で標準 `net/http` の `ServeMux` が **メソッド指定とパスパラメータ** に対応し、外部ルーターの必要性が大きく下がりました。

```go
mux.HandleFunc("GET /books/{id}", getBook)
id := r.PathValue("id")
```

| ライブラリ | 採用 OSS (調査対象のうち) | 特徴 |
| --- | --- | --- |
| 標準 `net/http` | Kubernetes, Prometheus, ほぼ全て | まずはこれ。第15章はこれだけで書く |
| [`go-chi/chi`](https://github.com/go-chi/chi) | Gitea, Caddy | `net/http` 互換の軽量ルーター。ミドルウェアが豊富 |
| [`gorilla/mux`](https://github.com/gorilla/mux) | Traefik (フォーク版) | 老舗。一時アーカイブされたが復活 |
| [`gin-gonic/gin`](https://github.com/gin-gonic/gin) | (Web アプリ・社内 API で人気) | 独自 Context。高速・多機能。GitHub スター数は最多 |
| [`labstack/echo`](https://github.com/labstack/echo) | (同上) | gin と並ぶフルスタック寄りフレームワーク |

**選び方の目安**: 標準 `net/http` → 足りなければ `chi` (標準互換を保てる) → チームの好みで gin/echo。

### API 定義駆動

- **OpenAPI**: [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen), [`ogen`](https://github.com/ogen-go/ogen) で OpenAPI 定義からサーバー/クライアントを生成
- **gRPC**: [`google.golang.org/grpc`](https://github.com/grpc/grpc-go) — etcd, Kubernetes, containerd, Terraform (プラグイン通信), Prometheus, k6, Traefik, Gitea, GitHub CLI が直接依存。マイクロサービス間通信の定番
- **Connect**: [`connectrpc/connect-go`](https://github.com/connectrpc/connect-go) — gRPC 互換かつ HTTP/1.1 + JSON でも呼べる

## 13.4 ロギング

Go 1.21 で **構造化ロギングの標準パッケージ `log/slog`** が追加されました。新規開発ではまず slog を検討しましょう。

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("request", "method", r.Method, "status", 200)
// {"time":"...","level":"INFO","msg":"request","method":"GET","status":200}
```

| ライブラリ | 採用 OSS (調査対象のうち) | 特徴 |
| --- | --- | --- |
| 標準 `log/slog` | (新規プロジェクトの第一候補) | 標準。Handler を差し替えて他ライブラリにも出力可能 |
| [`go.uber.org/zap`](https://github.com/uber-go/zap) | Kubernetes, etcd, Caddy, MinIO | 高速。slog 以前の構造化ログの定番 |
| [`rs/zerolog`](https://github.com/rs/zerolog) | Traefik | ゼロアロケーション |
| [`sirupsen/logrus`](https://github.com/sirupsen/logrus) | containerd, k6, lazygit, Traefik | 古参。現在はメンテナンスモード |
| `k8s.io/klog` | Kubernetes, containerd, Prometheus | Kubernetes エコシステム用 |

> 📦 Traefik の `go.mod` には `samber/slog-zerolog` (slog → zerolog のブリッジ) が含まれており、既存のロガーから slog へ移行していく過渡期の様子が見て取れます。

## 13.5 データベース

| ライブラリ | 種類 | 特徴 |
| --- | --- | --- |
| 標準 `database/sql` | 低レベル API | ドライバを差し替えて使う |
| [`jackc/pgx`](https://github.com/jackc/pgx) | PostgreSQL ドライバ | 高機能・高速。PostgreSQL ならこれ |
| [`sqlc-dev/sqlc`](https://github.com/sqlc-dev/sqlc) | SQL → Go コード生成 | **SQL を書くと型安全な Go 関数が生成される**。近年最も人気 |
| [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx) | database/sql 拡張 | 構造体へのスキャンが楽になる |
| [`gorm.io/gorm`](https://github.com/go-gorm/gorm) | ORM | Rails/Laravel 的な ORM。人気だが好みが分かれる |
| [`entgo.io/ent`](https://github.com/ent/ent) | ORM (コード生成) | Meta 発。スキーマを Go で定義 |
| [`xorm.io/xorm`](https://gitea.com/xorm/xorm) | ORM | Gitea が採用 |
| [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate), [`pressly/goose`](https://github.com/pressly/goose), [Atlas](https://atlasgo.io/) | マイグレーション | |

Go コミュニティでは **「ORM より SQL を直接書く (sqlc / pgx)」派が比較的多い** のが特徴です。第15章の Web API はインメモリ実装ですが、`Repository` インターフェースを sqlc + pgx の実装に差し替える発展課題を用意しています。

## 13.6 オブザーバビリティ

| ライブラリ | 採用 OSS (調査対象のうち) |
| --- | --- |
| [`prometheus/client_golang`](https://github.com/prometheus/client_golang) (メトリクス) | Prometheus, Caddy, containerd, Gitea, k6, MinIO, Traefik |
| [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) (トレース・メトリクス) | Kubernetes, Prometheus, Terraform, Caddy, containerd, k6, Traefik |

## 13.7 その他よく見るもの

| 用途 | ライブラリ |
| --- | --- |
| 並行処理ユーティリティ | `golang.org/x/sync` (errgroup, semaphore, singleflight) — 調査した 15 中 11 リポジトリが直接依存 |
| テスト | `stretchr/testify`, `google/go-cmp` (第11章) |
| UUID | `google/uuid` |
| バリデーション | `go-playground/validator` |
| HTTP クライアントのリトライ | `hashicorp/go-retryablehttp` |
| ファイル監視 | `fsnotify/fsnotify` |
| 埋め込み DB | `modernc.org/sqlite` (cgo 不要), `go.etcd.io/bbolt` |

## 13.8 ライブラリ選定の心得

1. **標準ライブラリでできないか先に確認** (*A little copying is better than a little dependency.*)
2. [pkg.go.dev](https://pkg.go.dev/) で **Imported by の数・最終更新・ライセンス** を確認
3. `go.mod` の依存が少ないライブラリを優先 (推移的依存も自分のリスクになる)
4. `govulncheck` で脆弱性を確認
5. [Awesome Go](https://github.com/avelino/awesome-go) はカタログとして便利

## 13.9 スタイルガイド

ライブラリと並んで、**コードの書き方** の公開ガイドも参考になります。

| ガイド | 発行元 | 特徴 |
| --- | --- | --- |
| [Effective Go](https://go.dev/doc/effective_go) | Go チーム | 原典。必読 |
| [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) | Go チーム | レビューでよく指摘される点のまとめ |
| [Google Go Style Guide](https://google.github.io/styleguide/go/) | Google | Guide / Decisions / Best Practices の 3 部構成。最も網羅的 |
| [Uber Go Style Guide](https://github.com/uber-go/guide) | Uber | 実践的な Do/Don't 集。日本語訳あり |
