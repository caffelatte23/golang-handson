# Go チュートリアル & ハンズオンブック

Go 言語の基礎から、**GitHub 上の主要 OSS の書き方・技術構成の調査**、そして **CLI / Web API / 並行処理ツール / ライブラリ** を実際に作るハンズオンまでをまとめた日本語の学習用ブックです。

- 対象: Go 初心者〜初中級者 (他言語の経験がある方)
- 動作確認: Go 1.24
- OSS 調査時点: 2026 年 11 月 (kubernetes, cli/cli, prometheus, terraform, hugo, caddy, traefik, gitea, etcd, containerd, k6, minio, fzf, lazygit, glow の 15 リポジトリ)

## 目次

### 第1部 基礎編

| 章 | 内容 |
| --- | --- |
| [第0章 はじめに](book/00-introduction.md) | Go の特徴・設計思想・Go で書かれた代表的 OSS |
| [第1章 環境構築と最初のプログラム](book/01-setup.md) | インストール、go コマンド、モジュール、クロスコンパイル |
| [第2章 基本文法](book/02-basics.md) | 変数・ゼロ値・制御構文・スライス/マップ・構造体・Functional Options |
| [第3章 インターフェースとジェネリクス](book/03-interfaces-generics.md) | 暗黙的インターフェース、io.Reader/Writer、ジェネリクス、イテレータ |
| [第4章 エラーハンドリング](book/04-errors.md) | `%w` ラップ、`errors.Is/As`、panic の使いどころ |
| [第5章 並行処理](book/05-concurrency.md) | goroutine、channel、mutex、context、errgroup、落とし穴 |

### 第2部 実践編 (OSS 調査)

| 章 | 内容 |
| --- | --- |
| [第6章 パッケージ設計とプロジェクト構成](book/06-project-layout.md) | 公式レイアウト、`internal/`、`pkg/` 論争、**主要 OSS 10 件の構成比較** |
| [第7章 テスト](book/07-testing.md) | テーブル駆動、httptest、Example/ベンチ/ファズ、**OSS のテストライブラリ採用状況** |
| [第8章 ツールチェイン・Lint・CI/CD](book/08-tooling-ci.md) | gofmt、golangci-lint (**cli/cli の設定解説**)、Makefile、GitHub Actions、Dependabot |
| [第9章 エコシステムと定番ライブラリ](book/09-ecosystem.md) | **15 OSS の go.mod 調査** に基づく CLI/HTTP/ログ/DB/監視ライブラリの選び方 |

### 第3部 ハンズオン

| 章 | 作るもの | 主な技術 | コード |
| --- | --- | --- | --- |
| [第10章 CLI ツール](book/10-handson-cli.md) | TODO 管理 CLI | cobra, JSON 永続化, 依存注入, tabwriter | [`handson/todo`](handson/todo) |
| [第11章 Web API](book/11-handson-webapi.md) | 書籍管理 REST API | net/http (1.22 ルーティング), slog, ミドルウェア, グレースフルシャットダウン | [`handson/bookshelf-api`](handson/bookshelf-api) |
| [第12章 並行処理](book/12-handson-concurrency.md) | 並行リンクチェッカー | goroutine, errgroup, context, 標準 flag | [`handson/linkcheck`](handson/linkcheck) |
| [第13章 ライブラリ](book/13-handson-library.md) | ジェネリック LRU キャッシュ | generics, iter.Seq2, Example, ベンチマーク, ファズ | [`handson/lru`](handson/lru) |
| [第14章 コンテナ化とリリース](book/14-container-release.md) | Docker イメージ / バイナリ配布 | マルチステージビルド, distroless, GoReleaser | — |

[付録 次に読むべき資料](book/99-appendix.md)

## リポジトリ構成

```
.
├── book/                 # 本文 (Markdown)
├── handson/              # ハンズオンの完成コード (それぞれ独立した Go モジュール)
│   ├── todo/
│   ├── bookshelf-api/
│   ├── linkcheck/
│   └── lru/
├── .github/workflows/    # CI (vet / test -race / golangci-lint)
├── .golangci.yml         # golangci-lint v2 設定
└── Makefile
```

## 完成コードを動かす

```bash
git clone https://github.com/caffelatte23/golang-hanson.git
cd golang-hanson

make test   # 全ハンズオンのテスト (-race 付き)
make lint   # golangci-lint (要インストール)

# 個別に動かす
cd handson/todo          && go run . add "Go を学ぶ" && go run . list
cd handson/bookshelf-api && go run ./cmd/server         # → http://localhost:8080/books
cd handson/linkcheck     && go run . https://go.dev
cd handson/lru           && go test -bench=. -benchmem
```

## 進め方のおすすめ

- **Go がはじめて**: 第0章から順に。各章の練習問題を解きつつ、第10章以降のハンズオンへ
- **他言語の経験が豊富**: 第2〜5章を流し読み → 第6〜9章で Go の「流儀」を把握 → ハンズオン
- **とにかく作りたい**: 第10章 or 第11章から。分からない文法が出たら第1部に戻る

ハンズオンは完成コードを写経するだけでなく、**各章末の「🚀 発展課題」に取り組む** ことで理解が深まります。
