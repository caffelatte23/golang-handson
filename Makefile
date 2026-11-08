# 各ハンズオンは独立した Go モジュールなので、ディレクトリごとにコマンドを実行します。
MODULES := $(wildcard handson/*/go.mod)
DIRS    := $(dir $(MODULES))

.PHONY: all test vet lint fmt tidy

all: fmt vet lint test

test: ## すべてのモジュールのテストを -race 付きで実行
	@for d in $(DIRS); do echo "==> $$d"; (cd $$d && go test -race ./...) || exit 1; done

vet:
	@for d in $(DIRS); do (cd $$d && go vet ./...) || exit 1; done

lint: ## golangci-lint (https://golangci-lint.run/) が必要
	@for d in $(DIRS); do echo "==> $$d"; (cd $$d && golangci-lint run --config $(CURDIR)/.golangci.yml ./...) || exit 1; done

fmt:
	@gofmt -l -w handson

tidy:
	@for d in $(DIRS); do (cd $$d && go mod tidy) || exit 1; done
