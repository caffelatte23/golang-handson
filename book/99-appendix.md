# 付録

## A Tour of Go / Effective Go との対応表

### A Tour of Go

| Tour のセクション | 本書 | 演習 |
| --- | --- | --- |
| Basics: Packages, variables, and functions | 第2章 | — |
| Basics: Flow control statements (for, if, else, switch and defer) | 第3章 | Exercise: Loops and Functions → [`exercises/sqrt`](../exercises/sqrt/) |
| Basics: More types (structs, slices, and maps) | 第4章 | Exercise: Slices → [`pic`](../exercises/pic/), Maps → [`wordcount`](../exercises/wordcount/), Fibonacci closure → [`fibonacci`](../exercises/fibonacci/) |
| Methods and interfaces | 第5章 | Stringers → [`stringer`](../exercises/stringer/), Readers → [`reader`](../exercises/reader/), rot13Reader → [`rot13`](../exercises/rot13/), Images → [`image`](../exercises/image/) |
| Methods and interfaces: Errors | 第7章 | Exercise: Errors → [`errsqrt`](../exercises/errsqrt/) |
| Generics | 第6章 | 本書独自 → [`generics`](../exercises/generics/) |
| Concurrency | 第8章 | Equivalent Binary Trees → [`equivtree`](../exercises/equivtree/), Web Crawler → [`crawler`](../exercises/crawler/) |

### Effective Go

| Effective Go の節 | 本書 |
| --- | --- |
| Formatting | 9.1 |
| Commentary | 9.2 |
| Names (Package names, Getters, Interface names, MixedCaps) | 9.3 |
| Semicolons | 2.12 |
| Control structures (If, Redeclaration and reassignment, For, Switch, Type switch) | 3.1〜3.3, 2.5, 5.4 |
| Functions (Multiple return values, Named result parameters, Defer) | 2.4, 3.4 |
| Data (new, Constructors and composite literals, make, Arrays, Slices, Two-dimensional slices, Maps, Printing, Append) | 第4章 |
| Initialization (Constants, Variables, The init function) | 2.10, 2.11 |
| Methods (Pointers vs. Values) | 5.1, 5.2 |
| Interfaces and other types (Interfaces, Conversions, Interface conversions and type assertions, Generality, Interfaces and methods) | 5.3〜5.6 |
| The blank identifier | 9.4, 5.4 |
| Embedding | 5.7 |
| Concurrency (Share by communicating, Goroutines, Channels, Channels of channels, Parallelization, A leaky buffer) | 8.1〜8.6 |
| Errors (Panic, Recover) | 第7章 |
| A web server | 9.6 |

Effective Go は 2009 年に書かれて以降、大きくは更新されていません。そのため、ジェネリクス、イテレータ、モジュール、`errors.Is` / `errors.As`、`context`、`log/slog` といった比較的新しい機能は扱っていません。本書ではこれらを第6〜8章と第2部で補っています。

## 次に読むべき資料

### 公式

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

### スタイルガイド

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide)

### 書籍

- *The Go Programming Language* (Donovan & Kernighan) — 邦訳『プログラミング言語Go』
- *Learning Go, 2nd Edition* (Jon Bodner) — 邦訳『初めてのGo言語 第2版』
- *100 Go Mistakes and How to Avoid Them* (Teiva Harsanyi) — よくある間違い 100 選
- *Concurrency in Go* (Katherine Cox-Buday) — 邦訳『Go言語による並行処理』

### 読むと勉強になる OSS (難易度順)

| OSS | 規模 | 学べること |
| --- | --- | --- |
| [junegunn/fzf](https://github.com/junegunn/fzf) | 小〜中 | 外部依存の少ない CLI、ターミナル制御 |
| [charmbracelet/glow](https://github.com/charmbracelet/glow) | 小 | cobra + viper + bubbletea の典型構成 |
| [cli/cli](https://github.com/cli/cli) | 中 | cobra の大規模運用、テスト容易な設計、GoReleaser |
| [caddyserver/caddy](https://github.com/caddyserver/caddy) | 中 | プラグイン (モジュール) アーキテクチャ、`net/http` の深い活用 |
| [prometheus/prometheus](https://github.com/prometheus/prometheus) | 大 | 時系列 DB、パーサー、並行処理 |
| [hashicorp/terraform](https://github.com/hashicorp/terraform) | 大 | `internal/` 中心の構成、gRPC プラグイン |
| [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes) | 特大 | コントローラーパターン、コード生成、モノレポ運用 |

### コミュニティ

- [Gophers Slack](https://invite.slack.golangbridge.org/) (`#japan` チャンネルあり)
- [Go Conference](https://gocon.jp/) — 日本の Go カンファレンス
- [Go Forum](https://forum.golangbridge.org/)

## クレジット

- 第1部の構成とトピック、演習のテーマは、Go Authors による [A Tour of Go](https://go.dev/tour/) と [Effective Go](https://go.dev/doc/effective_go) を参考にしています。go.dev のドキュメントは [Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/) で、コードは [BSD ライセンス](https://go.dev/LICENSE) で公開されています。本書の解説とサンプルコードは本書のために書き下ろしたものです。ただし、`ByteSize`、`trace`/`un`、`safelyDo`、`Sequence`、`Job`、リークするバッファなど、Effective Go の有名な例を元にしたものはその旨を本文に記しています。
- 「Go Proverbs」は Rob Pike の講演 (GopherFest 2015) によるものです。
- 第2部の OSS 調査では、各プロジェクトの公開リポジトリにある `go.mod` や設定ファイルを参照しました。引用したコードの著作権は、それぞれのプロジェクトに帰属します。
