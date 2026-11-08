# 第0章 はじめに — Go とはどんな言語か

## この本のねらい

本書は、Go を書いたことがない人や少しだけ触ったことがある人が、**この 1 冊で Go の基礎から実務の入り口までをひととおり学べる** ように作った日本語のハンズオンブックです。

1. **言語の基礎**: 公式チュートリアル [A Tour of Go](https://go.dev/tour/) と公式ガイド [Effective Go](https://go.dev/doc/effective_go) が扱う範囲をすべて含めています。最新の言語仕様 (ジェネリクス、イテレータなど) と、現場でよく使う書き方も加えています。
2. **OSS の調査**: GitHub 上の主要な OSS が、どんな構成・ツール・ライブラリで書かれているかを調べて紹介します。
3. **ハンズオン**: **CLI・Web API・並行処理ツール・ライブラリ** を実際に作って動かします。

第2部の「📦 OSS ではこう書かれている」コラムでは、実際の OSS リポジトリ (2026 年 11 月時点の `go.mod` や設定ファイル) を調べた結果を紹介します。

### A Tour of Go / Effective Go との関係

本書の第1部のうち第2〜9章は、この 2 つの公式ドキュメントで扱っているトピックを網羅しています。説明は本書のために書き下ろしたもので、Go 1.24 以降の言語仕様と標準ライブラリに合わせて内容を補っています。サンプルコードの一部は Tour と Effective Go のプログラムを元にしています (一覧は付録のクレジットを参照)。

| 公式ドキュメント | 本書での扱い |
| --- | --- |
| A Tour of Go の解説 | 第2〜8章の本文。そのまま実行できるサンプルプログラムは、すべて CI で実行して出力を確かめています |
| A Tour of Go の演習 (Exercise) | [`exercises/`](../exercises/) に、**自動で採点されるテスト付き** で収録しています。解答例も付いています |
| Effective Go | 各トピックは関係する章に組み込み、書き方の約束事は第9章にまとめています |

章ごとの詳しい対応は、[付録](99-appendix.md#a-tour-of-go--effective-go-との対応表) を参照してください。

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
第1部 基礎編 (A Tour of Go + Effective Go)
  第1章 環境構築と最初のプログラム
  第2章 基本 — パッケージ・関数・変数・型・定数
  第3章 制御構文 — for・if・switch・defer
  第4章 データ構造 — ポインタ・構造体・配列・スライス・マップ・関数値
  第5章 メソッドとインターフェース
  第6章 ジェネリクスとイテレータ
  第7章 エラーハンドリング
  第8章 並行処理
  第9章 Go らしいコードの書き方 (Effective Go)
第2部 実践編 (OSS の調査)
  第10章 パッケージ設計とプロジェクト構成
  第11章 テスト
  第12章 ツールチェイン・Lint・CI/CD・リリース
  第13章 エコシステムと定番ライブラリ
第3部 ハンズオン
  第14章 CLI ツール: TODO 管理 (cobra)
  第15章 Web API: 書籍管理 REST API (net/http + slog)
  第16章 並行処理: リンクチェッカー (goroutine + errgroup)
  第17章 ライブラリ: ジェネリックな LRU キャッシュ (ジェネリクス + イテレータ + ファズテスト)
  第18章 コンテナ化とリリース
付録 Tour / Effective Go との対応表、参考資料、クレジット
```

### 演習問題の進め方

第3〜8章の終わりには、A Tour of Go と同じテーマの演習があります。リポジトリを clone して、`exercises/` ディレクトリで取り組みます。

```bash
git clone https://github.com/caffelatte23/golang-hanson.git
cd golang-hanson/exercises
# exercise.go の TODO を実装して、テストが通れば正解
go test ./sqrt/
# 解答例のコードでテストを実行する
go test -tags solution ./sqrt/
```

基礎に自信がある人は第2部から、とにかく手を動かしたい人は第3部から読み始めても構いません。

## 前提

- Go **1.24 以上** (本書のサンプルは Go 1.24 と Go 1.27 で動作を確認しています)。Go は最新の 2 つのメジャーリリースだけがセキュリティ修正などのサポート対象なので、**サポートされている最新版** を使ってください。第1章のインストール手順のバージョン番号も、[公式ダウンロードページ](https://go.dev/dl/) の最新版に読み替えてください。
- Git、ターミナルの基本操作
- 何らかのプログラミング言語の経験
