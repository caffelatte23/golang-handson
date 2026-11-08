# 付録 次に読むべき資料

## 公式

| 資料 | 内容 |
| --- | --- |
| [A Tour of Go](https://go.dev/tour/) | ブラウザで動く公式チュートリアル (日本語版あり) |
| [Effective Go](https://go.dev/doc/effective_go) | Go らしい書き方の原典 |
| [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) | コードレビューでの頻出指摘 |
| [Organizing a Go module](https://go.dev/doc/modules/layout) | 公式のプロジェクト構成ガイド |
| [The Go Blog](https://go.dev/blog/) | 新機能の解説・設計思想 |
| [Go Release Notes](https://go.dev/doc/devel/release) | バージョンごとの変更点 |
| [Tutorial: Developing a RESTful API with Go and Gin](https://go.dev/doc/tutorial/web-service-gin) | 公式の Web API チュートリアル |
| [Go by Example](https://gobyexample.com/) | 短いサンプル集 (非公式だが定番) |

## スタイルガイド

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide)

## 書籍

- *The Go Programming Language* (Donovan & Kernighan) — 邦訳『プログラミング言語Go』
- *Learning Go, 2nd Edition* (Jon Bodner) — 邦訳『初めてのGo言語 第2版』
- *100 Go Mistakes and How to Avoid Them* (Teiva Harsanyi) — よくある間違い 100 選
- *Concurrency in Go* (Katherine Cox-Buday) — 邦訳『Go言語による並行処理』

## 読むと勉強になる OSS (難易度順)

| OSS | 規模 | 学べること |
| --- | --- | --- |
| [junegunn/fzf](https://github.com/junegunn/fzf) | 小〜中 | 外部依存の少ない CLI、ターミナル制御 |
| [charmbracelet/glow](https://github.com/charmbracelet/glow) | 小 | cobra + viper + bubbletea の典型構成 |
| [cli/cli](https://github.com/cli/cli) | 中 | cobra の大規模運用、テスト容易な設計、GoReleaser |
| [caddyserver/caddy](https://github.com/caddyserver/caddy) | 中 | プラグイン (モジュール) アーキテクチャ、`net/http` の深い活用 |
| [prometheus/prometheus](https://github.com/prometheus/prometheus) | 大 | 時系列 DB、パーサー、並行処理 |
| [hashicorp/terraform](https://github.com/hashicorp/terraform) | 大 | `internal/` 中心の構成、gRPC プラグイン |
| [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes) | 特大 | コントローラーパターン、コード生成、モノレポ運用 |

## コミュニティ

- [Gophers Slack](https://invite.slack.golangbridge.org/) (`#japan` チャンネルあり)
- [Go Conference](https://gocon.jp/) — 日本の Go カンファレンス
- [Go Forum](https://forum.golangbridge.org/)
