# 第18章 コンテナ化とリリース

作ったアプリを **届ける** ところまでやってみましょう。Web API はコンテナイメージとして、CLI は複数 OS 向けバイナリとして配布するのが一般的です。

## 18.1 Web API をコンテナ化する

`handson/bookshelf-api/Dockerfile`:

```dockerfile
# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.24 AS build
WORKDIR /src
# 依存だけ先にダウンロードしてレイヤーキャッシュを効かせる
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- runtime stage ----
# distroless: シェルもパッケージマネージャもない最小イメージ
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

**ポイント**

- **マルチステージビルド**: ビルド用イメージ (`golang`、数百 MB〜1GB 程度) でコンパイルし、実行用イメージには **バイナリだけ** をコピーします。
- **`CGO_ENABLED=0`**: C ライブラリに依存しない静的バイナリにすることで、libc すら無い `distroless/static` や `scratch` で動きます。
- **`-trimpath`**: バイナリからビルドマシンのファイルパスを除去 (再現性・情報漏えい対策)。
- **`-ldflags="-s -w"`**: シンボル・デバッグ情報を削除してサイズ削減。
- **`go.mod` / `go.sum` を先にコピーして `go mod download`**: ソースだけ変更した場合に依存ダウンロードのレイヤーキャッシュが効きます (依存がある場合は `COPY go.mod go.sum ./`)。
- **nonroot ユーザー** で実行。

最終イメージは **おおむね 10MB 前後** になります (このサンプルの Go バイナリは約 6MB で、distroless/static のベースは数 MB です)。Node.js や Python の API と比べて桁違いに小さく、起動も一瞬です。

```bash
cd handson/bookshelf-api
docker build -t bookshelf-api .
docker run --rm -p 8080:8080 bookshelf-api
docker images bookshelf-api   # サイズを確認
```

> 📦 **OSS ではこう書かれている**: 調査対象の Prometheus, Terraform, Traefik, Gitea, fzf, Glow, etcd, k6, MinIO, lazygit, Hugo はリポジトリ直下に `Dockerfile` を持っています。Kubernetes のコンポーネントも distroless ベースのイメージで配布されています。

### ko — Dockerfile なしでイメージを作る

[ko](https://ko.build/) は Go 専用のコンテナイメージビルダーです。Dockerfile も Docker デーモンも不要で、`ko build ./cmd/server` だけで distroless ベースのイメージを作ってレジストリへ push できます (Knative や Sigstore などが採用)。

## 18.2 CLI を GoReleaser でリリースする

`handson/todo/.goreleaser.yaml`:

```yaml
# GoReleaser v2 設定 (https://goreleaser.com/)
# ローカルで試す: goreleaser release --snapshot --clean
version: 2

project_name: todo

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - main: .
    binary: todo
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    flags:
      - -trimpath
    ldflags:
      - -s -w -X github.com/caffelatte23/golang-hanson/handson/todo/internal/cli.version={{.Version}}

archives:
  - formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

```bash
# インストール
go install github.com/goreleaser/goreleaser/v2@latest

# 設定チェック
goreleaser check

# タグを打たずにローカルで全ターゲットをビルド (dist/ に出力)
goreleaser release --snapshot --clean
ls dist/
```

### GitHub Actions でタグ push 時に自動リリース

`.github/workflows/release.yml` (**go.mod がリポジトリ直下にある** 単一モジュールのリポジトリに置く場合の例です。本リポジトリの `handson/todo` のようにサブディレクトリにある場合は、`goreleaser-action` に `workdir: handson/todo` を、`setup-go` に `go-version-file: handson/todo/go.mod` を指定します):

```yaml
name: release
on:
  push:
    tags: ["v*"]

permissions:
  contents: write   # GitHub Release の作成に必要

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # changelog 生成に全履歴が必要
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

```bash
git tag v0.1.0 && git push origin v0.1.0
# → GitHub Releases に linux/darwin/windows × amd64/arm64 のバイナリが並ぶ
```

さらに `brews:` (Homebrew tap)、`nfpms:` (deb/rpm)、`dockers:` (コンテナイメージ)、`signs:` (cosign 署名) などを追加すれば、パッケージマネージャ経由の配布も自動化できます。GitHub CLI の [`.goreleaser.yml`](https://github.com/cli/cli/blob/trunk/.goreleaser.yml) はその全部入りの好例です。

## 18.3 配布方法のまとめ

| 方法 | 対象 | コマンド |
| --- | --- | --- |
| `go install` | Go 開発者 | `go install github.com/you/todo@latest` |
| GitHub Releases | 全員 | バイナリをダウンロード |
| Homebrew | macOS/Linux ユーザー | `brew install you/tap/todo` |
| コンテナイメージ | サーバー | `docker run ghcr.io/you/bookshelf-api` |
| deb / rpm / winget / scoop | 各 OS ユーザー | GoReleaser で生成 |

## ✅ チェックリスト

- [ ] `CGO_ENABLED=0` で静的バイナリがビルドできる
- [ ] Docker イメージが distroless / scratch ベースで小さい
- [ ] バージョン情報が `--version` で表示される
- [ ] タグ push で GitHub Release が自動作成される
