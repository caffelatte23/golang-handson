# 第1章 環境構築と最初のプログラム

## 1.1 インストール

[公式ダウンロードページ](https://go.dev/dl/) からインストーラーを取得するのが最も確実です。

```bash
# macOS (Homebrew)
brew install go

# Linux (公式 tarball)
curl -LO https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc

# 確認
go version
```

> 💡 **複数バージョンの管理**: Go 1.21 以降は `go.mod` の `go` / `toolchain` 行に応じて必要なツールチェインを自動ダウンロードします (`GOTOOLCHAIN` 環境変数で制御)。そのため `goenv` などのバージョンマネージャは必須ではありません。

### `go env` で設定を確認

```bash
go env GOPATH GOROOT GOMODCACHE GOOS GOARCH GOTOOLCHAIN
```

| 変数 | 意味 |
| --- | --- |
| `GOROOT` | Go 本体のインストール先 |
| `GOPATH` | `go install` したバイナリ (`$GOPATH/bin`) やモジュールキャッシュの置き場 (既定 `~/go`) |
| `GOOS` / `GOARCH` | ビルド対象の OS / CPU。変えるだけでクロスコンパイルできる |

## 1.2 エディタ

公式の言語サーバー **gopls** を使うのが標準です。

- **VS Code**: 拡張機能「Go」(Go チーム公式) をインストール → コマンドパレットで `Go: Install/Update Tools`
- **GoLand**: JetBrains の Go 専用 IDE
- **Neovim / Emacs / Helix**: gopls を LSP クライアントで利用

保存時に `gofmt` (または `goimports`) が走るように設定しておきましょう。Go のコードは **全員が同じフォーマット** で書くのが文化です。

## 1.3 Hello, World

```bash
mkdir hello && cd hello
go mod init example.com/hello   # モジュールを初期化 (go.mod が作られる)
```

`main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, 世界")
}
```

```bash
go run .            # コンパイルして即実行
go build -o hello . # 実行ファイルを生成
./hello
```

### モジュールパスの付け方

`go mod init` に渡すのは **モジュールパス** です。公開するなら GitHub の URL と一致させます。

```bash
go mod init github.com/<your-name>/<repo>
```

OSS の例:

| リポジトリ | モジュールパス |
| --- | --- |
| cli/cli | `github.com/cli/cli/v2` (メジャーバージョン 2 以上は `/v2` を付ける) |
| kubernetes/kubernetes | `k8s.io/kubernetes` (独自ドメイン = vanity import path) |
| hashicorp/terraform | `github.com/hashicorp/terraform` |

## 1.4 よく使う go コマンド

| コマンド | 用途 |
| --- | --- |
| `go run .` | ビルドして実行 |
| `go build ./...` | 全パッケージをビルド (`./...` は配下すべての意味) |
| `go test ./...` | テスト実行 |
| `go vet ./...` | 静的解析 (バグの可能性を検出) |
| `go fmt ./...` / `gofmt -w .` | フォーマット |
| `go mod tidy` | `go.mod` / `go.sum` を import 文に合わせて整理 |
| `go get pkg@version` | 依存の追加・更新 |
| `go install pkg@version` | ツールを `$GOPATH/bin` にインストール |
| `go doc fmt.Println` | ドキュメントをターミナルで表示 |
| `go tool <name>` | `go.mod` の `tool` ディレクティブで管理したツールを実行 (Go 1.24+) |

## 1.5 クロスコンパイル

```bash
GOOS=windows GOARCH=amd64 go build -o hello.exe .
GOOS=darwin  GOARCH=arm64 go build -o hello-mac .
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o hello-linux-arm64 .
```

`CGO_ENABLED=0` にすると C ライブラリに依存しない完全な静的バイナリになり、`scratch` や `distroless` といった極小コンテナイメージでも動きます (第18章)。

## 1.6 Playground

ブラウザだけで試したいときは [Go Playground](https://go.dev/play/) が便利です。公式チュートリアル [A Tour of Go](https://go.dev/tour/) もここで動きます。

## ✅ チェックリスト

- [ ] `go version` が 1.24 以上を表示する
- [ ] エディタで保存時に自動フォーマットされる
- [ ] `go run .` で Hello, World が表示できた
