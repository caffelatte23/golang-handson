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
