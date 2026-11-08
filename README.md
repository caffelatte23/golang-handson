# Go チュートリアル & ハンズオンブック

Go 言語の基礎から、**GitHub 上の主要 OSS の書き方・技術構成の調査**、そして **CLI / Web API / 並行処理ツール / ライブラリ** を実際に作るハンズオンまでをまとめた日本語の学習用ブックです。

第1部は、公式の [A Tour of Go](https://go.dev/tour/) と [Effective Go](https://go.dev/doc/effective_go) が扱う内容をすべて含んでいます。本書だけで、言語の基礎から Go らしい書き方まで学べます。本文のサンプルプログラムは CI で実際に実行し、掲載している出力と一致することを確かめています。Tour と同じテーマの演習には、自動採点のテストを付けています。

- 対象: Go 初心者〜初中級者 (他言語の経験がある方)
- 動作確認: Go 1.24
- OSS 調査時点: 2026 年 11 月 (kubernetes, cli/cli, prometheus, terraform, hugo, caddy, traefik, gitea, etcd, containerd, k6, minio, fzf, lazygit, glow の 15 リポジトリ)

## 目次

### 第1部 基礎編 (A Tour of Go + Effective Go)

| 章 | 内容 | 演習 |
| --- | --- | --- |
| [第0章 はじめに](book/00-introduction.md) | Go の特徴・設計思想・Go で書かれた代表的な OSS | |
| [第1章 環境構築と最初のプログラム](book/01-setup.md) | インストール、go コマンド、モジュール、クロスコンパイル | |
| [第2章 基本](book/02-basics.md) | パッケージ、関数、変数、基本型、ゼロ値、型変換、定数、iota、init | |
| [第3章 制御構文](book/03-flow-control.md) | for、range、if、switch、defer | 平方根 |
| [第4章 データ構造](book/04-data-structures.md) | ポインタ、構造体、配列、スライス、マップ、new と make、クロージャ、fmt | Pic / WordCount / Fibonacci |
| [第5章 メソッドとインターフェース](book/05-methods-interfaces.md) | メソッド、ポインタレシーバ、インターフェース値、型 switch、Stringer、io.Reader、埋め込み | Stringer / Reader / rot13 / Image |
| [第6章 ジェネリクスとイテレータ](book/06-generics.md) | 型パラメータ、制約、ジェネリックな型、iter.Seq、iter.Pull | Map / Filter / List |
| [第7章 エラーハンドリング](book/07-errors.md) | error 型、`%w`、`errors.Is` / `errors.As`、panic、recover | ErrNegativeSqrt |
| [第8章 並行処理](book/08-concurrency.md) | goroutine、channel、select、sync、Effective Go のパターン、context、errgroup | 二分木 / Web クローラー |
| [第9章 Go らしいコードの書き方](book/09-effective-go.md) | gofmt、doc comment、命名、空白識別子、Functional Options、Go Proverbs | |

### 第2部 実践編 (OSS 調査)

| 章 | 内容 |
| --- | --- |
| [第10章 パッケージ設計とプロジェクト構成](book/10-project-layout.md) | 公式レイアウト、`internal/`、`pkg/` 論争、**主要 OSS 10 件の構成比較** |
| [第11章 テスト](book/11-testing.md) | テーブル駆動、httptest、Example/ベンチ/ファズ、**OSS のテストライブラリ採用状況** |
| [第12章 ツールチェイン・Lint・CI/CD](book/12-tooling-ci.md) | gofmt、golangci-lint (**cli/cli の設定解説**)、Makefile、GitHub Actions、Dependabot |
| [第13章 エコシステムと定番ライブラリ](book/13-ecosystem.md) | **15 OSS の go.mod 調査** に基づく CLI/HTTP/ログ/DB/監視ライブラリの選び方 |

### 第3部 ハンズオン

| 章 | 作るもの | 主な技術 | コード |
| --- | --- | --- | --- |
| [第14章 CLI ツール](book/14-handson-cli.md) | TODO 管理 CLI | cobra, JSON 永続化, 依存注入, tabwriter | [`handson/todo`](handson/todo) |
| [第15章 Web API](book/15-handson-webapi.md) | 書籍管理 REST API | net/http (1.22 ルーティング), slog, ミドルウェア, グレースフルシャットダウン | [`handson/bookshelf-api`](handson/bookshelf-api) |
| [第16章 並行処理](book/16-handson-concurrency.md) | 並行リンクチェッカー | goroutine, errgroup, context, 標準 flag | [`handson/linkcheck`](handson/linkcheck) |
| [第17章 ライブラリ](book/17-handson-library.md) | ジェネリック LRU キャッシュ | generics, iter.Seq2, Example, ベンチマーク, ファズ | [`handson/lru`](handson/lru) |
| [第18章 コンテナ化とリリース](book/18-container-release.md) | Docker イメージ / バイナリ配布 | マルチステージビルド, distroless, GoReleaser | — |

[付録 Tour / Effective Go との対応表・参考資料・クレジット](book/99-appendix.md)

## リポジトリ構成

```
.
├── book/                 # 本文 (Markdown)
├── exercises/            # 第3〜8章の演習 (自動採点のテストと解答例付き)
├── handson/              # ハンズオンの完成コード (それぞれ独立した Go モジュール)
│   ├── todo/
│   ├── bookshelf-api/
│   ├── linkcheck/
│   └── lru/
├── scripts/              # 本文のサンプルコードを検証するスクリプト
├── .github/workflows/    # CI (本文の検証 / 演習の解答 / ハンズオンの test・lint)
├── .golangci.yml         # golangci-lint v2 設定
└── Makefile
```

## 完成コードを動かす

```bash
git clone https://github.com/caffelatte23/golang-hanson.git
cd golang-hanson

make test   # 全ハンズオンと演習の解答例のテスト (-race 付き)
make book   # 本文のサンプルプログラムを実行して、掲載している出力と照合する
make lint   # golangci-lint (要インストール)

# 個別に動かす
cd handson/todo          && go run . add "Go を学ぶ" && go run . list
cd handson/bookshelf-api && go run ./cmd/server         # → http://localhost:8080/books
cd handson/linkcheck     && go run . https://go.dev
cd handson/lru           && go test -bench=. -benchmem
```

## 進め方のおすすめ

- **Go がはじめて**: 第0章から順に読み、各章の演習 (`exercises/`) を解いてから、第14章以降のハンズオンに進みます。
- **他の言語の経験が豊富**: 第2〜8章は流し読みでかまいません。第9章で Go の流儀を、第10〜13章で OSS の実践を押さえてから、ハンズオンに進みます。
- **とにかく作りたい**: 第14章か第15章から始め、分からない文法が出てきたら第1部に戻ります。

ハンズオンは完成コードを写経するだけでなく、**各章末の「🚀 発展課題」に取り組む** ことで理解が深まります。

## ライセンス

| 対象 | ライセンス |
| --- | --- |
| 本文 (`book/`、各 README などのドキュメント) | [CC BY 4.0](LICENSE-CC-BY-4.0) |
| コード (`exercises/`、`handson/`、`scripts/`、本文中のサンプルコード) | [MIT](LICENSE) |

第1部は Go Authors による A Tour of Go と Effective Go (CC BY 4.0 / BSD) を参考にしています。詳しくは [付録のクレジット](book/99-appendix.md#クレジット) を参照してください。
