# 第0章 はじめに — Go とはどんな言語か

## この本のねらい

本書は「Go を書いたことがない/少しだけ触った」エンジニアが、

1. 言語の基本と **Go らしい書き方 (idiomatic Go)** を身につけ、
2. GitHub 上の主要 OSS が **どんな構成・ツール・ライブラリで書かれているか** を知り、
3. **CLI / Web API / 並行処理ツール / ライブラリ** を実際に作って動かす

ところまでを一気通貫で体験することを目的としたハンズオンブックです。

各章の「📦 OSS ではこう書かれている」コラムでは、実際の OSS リポジトリ (2026 年 11 月時点の `go.mod` や設定ファイル) を調査した結果を紹介します。

## Go の特徴

| 特徴 | 内容 | 現場でうれしいこと |
| --- | --- | --- |
| シンプルな言語仕様 | キーワードは 25 個。クラス・継承・例外がない | 他人のコードが読みやすい。レビューしやすい |
| 静的型付け + 型推論 | `x := 1` のように書ける | 動的言語の書き味とコンパイル時チェックの両立 |
| 高速なコンパイル | 大規模でも数秒〜数十秒 | 開発ループが速い |
| 単一バイナリ | 依存を含めた実行ファイル 1 つを出力。クロスコンパイルも簡単 | 配布・デプロイ・コンテナ化が楽 |
| 並行処理が言語組み込み | goroutine と channel | サーバーや I/O 多重化が書きやすい |
| 充実した標準ライブラリ | HTTP サーバー/クライアント、JSON、暗号、テスト等 | 外部依存を最小化できる |
| 統一されたツールチェイン | `go build/test/vet/fmt/mod/doc` が標準装備 | フォーマット論争がない。CI が単純 |
| 後方互換性の約束 | [Go 1 互換性ポリシー](https://go.dev/doc/go1compat) | 10 年前のコードがそのまま動く |

### Go の設計思想 (Go Proverbs より)

Rob Pike による [Go Proverbs](https://go-proverbs.github.io/) は、Go コミュニティの価値観をよく表しています。本書でも繰り返し登場します。

- *Don't communicate by sharing memory, share memory by communicating.* — メモリ共有で通信するな、通信でメモリを共有せよ
- *The bigger the interface, the weaker the abstraction.* — インターフェースは小さく
- *Make the zero value useful.* — ゼロ値をそのまま使えるように設計する
- *Errors are values.* — エラーはただの値。普通のロジックで扱う
- *Clear is better than clever.* — 賢いコードより明快なコード
- *A little copying is better than a little dependency.* — 少しのコピーは少しの依存よりマシ

## Go が使われている代表的な OSS

Go は特に **クラウドインフラ・CLI・ネットワークサーバー** の分野で圧倒的な採用実績があります。

| 分野 | OSS | リポジトリ |
| --- | --- | --- |
| コンテナ/オーケストレーション | Kubernetes, containerd, Docker (Moby) | `kubernetes/kubernetes`, `containerd/containerd`, `moby/moby` |
| IaC | Terraform | `hashicorp/terraform` |
| 監視 | Prometheus, Grafana k6 | `prometheus/prometheus`, `grafana/k6` |
| 分散 KVS | etcd | `etcd-io/etcd` |
| Web サーバー/プロキシ | Caddy, Traefik | `caddyserver/caddy`, `traefik/traefik` |
| Git ホスティング | Gitea | `go-gitea/gitea` |
| オブジェクトストレージ | MinIO | `minio/minio` |
| CLI ツール | GitHub CLI (gh), fzf, lazygit, Glow | `cli/cli`, `junegunn/fzf`, `jesseduffield/lazygit`, `charmbracelet/glow` |
| 静的サイトジェネレーター | Hugo | `gohugoio/hugo` |

これらはすべて本書で「お手本」として参照します。ソースを手元に clone して読み比べるのは、Go 上達の最短ルートの一つです。

## 本書の構成

```
第1部 基礎編
  第1章 環境構築と最初のプログラム
  第2章 基本文法
  第3章 インターフェースとジェネリクス
  第4章 エラーハンドリング
  第5章 並行処理
第2部 実践編 (OSS 調査)
  第6章 パッケージ設計とプロジェクト構成
  第7章 テスト
  第8章 ツールチェイン・Lint・CI/CD・リリース
  第9章 エコシステムと定番ライブラリ
第3部 ハンズオン
  第10章 CLI ツール: TODO 管理 (cobra)
  第11章 Web API: 書籍管理 REST API (net/http + slog)
  第12章 並行処理: リンクチェッカー (goroutine + errgroup)
  第13章 ライブラリ: ジェネリック LRU キャッシュ (generics + iterator + fuzz)
  第14章 コンテナ化とリリース
付録 次に読むべき資料
```

基礎に自信がある人は第2部から、とにかく手を動かしたい人は第3部から読み始めても構いません。

## 前提

- Go **1.24 以上** (本書のサンプルは 1.24 で動作確認。新しいバージョンでもそのまま動きます)
- Git、ターミナルの基本操作
- 何らかのプログラミング言語の経験
