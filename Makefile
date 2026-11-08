# ハンズオンと演習は、それぞれ独立した Go モジュールなので、ディレクトリごとにコマンドを実行します。
MODULES := $(wildcard handson/*/go.mod)
DIRS    := $(dir $(MODULES))

.PHONY: help all test vet lint fmt tidy book exercises

all: fmt vet lint test book ## フォーマット・静的解析・テスト・本文の検証をすべて実行する

help: ## ターゲットの一覧を表示する
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

test: exercises ## ハンズオンをすべて -race 付きでテストする
	@for d in $(DIRS); do echo "==> $$d"; (cd $$d && go test -race ./...) || exit 1; done

exercises: ## 演習を解答例 (-tags solution) でテストする
	@echo "==> exercises (solution)"; cd exercises && go vet ./... && go test -race -tags solution ./...

vet: ## go vet を実行する
	@for d in $(DIRS); do (cd $$d && go vet ./...) || exit 1; done

lint: ## golangci-lint (https://golangci-lint.run/) が必要
	@for d in $(DIRS); do echo "==> $$d"; (cd $$d && golangci-lint run --config $(CURDIR)/.golangci.yml ./...) || exit 1; done
	@echo "==> exercises"; cd exercises && golangci-lint run --config $(CURDIR)/.golangci.yml --build-tags solution ./...

book: ## 本文のサンプルプログラムを実行し、掲載している出力と照合する
	@python3 scripts/check_book_code.py

fmt: ## gofmt で整形する
	@gofmt -l -w handson exercises

tidy: ## go mod tidy を実行する
	@for d in $(DIRS) exercises; do (cd $$d && go mod tidy) || exit 1; done
