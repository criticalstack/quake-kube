BIN_DIR ?= bin
LDFLAGS := -s -w
GOFLAGS = -gcflags "all=-trimpath=$(PWD)" -asmflags "all=-trimpath=$(PWD)"
GO_BUILD_ENV_VARS := GO111MODULE=on CGO_ENABLED=0

q3: gen
	@$(GO_BUILD_ENV_VARS) go build -o $(BIN_DIR)/q3 $(GOFLAGS) -ldflags '$(LDFLAGS)' ./cmd/q3

gen: ## Generate and embed templates
	@go run tools/genstatic.go public public

VERSION ?= v1.0.0
IMAGE   ?= docker.io/criticalstack/quake:$(VERSION)

.PHONY: build
build:
	@docker build . --force-rm --build-arg GOPROXY --build-arg GOSUMDB -t $(IMAGE)
