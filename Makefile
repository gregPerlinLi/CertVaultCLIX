BINARY := cvx
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X github.com/gregPerlinLi/CertVaultCLIX/internal/version.Version=$(VERSION) \
	-X github.com/gregPerlinLi/CertVaultCLIX/internal/version.Commit=$(COMMIT) \
	-X github.com/gregPerlinLi/CertVaultCLIX/internal/version.BuildDate=$(DATE)

.PHONY: build clean install test lint

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

clean:
	rm -f $(BINARY) $(BINARY)-*
	rm -rf dist/

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := build
