.DEFAULT_GOAL := build

# globals
BINARY_NAME?=btc-price-service
BUILD_DIR?="./build"
CGO_ENABLED?=0
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%S %Z')
REPO=github.com/gandarez/btc-price-service
VERSION?=<local-build>

# ld flags for go build
LD_FLAGS=-s -w -X '${REPO}/internal/foundation/version.BuildDate=${DATE}' -X ${REPO}/internal/foundation/version.Commit=${COMMIT} -X ${REPO}/internal/foundation/version.Version=${VERSION}

# basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# linting
define get_latest_lint_release
	curl -s "https://api.github.com/repos/golangci/golangci-lint/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
endef
LATEST_LINT_VERSION=$(shell $(call get_latest_lint_release))
INSTALLED_LINT_VERSION=$(shell golangci-lint --version 2>/dev/null | awk '{print "v"$$4}')

# targets
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(MAKE) build

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -v \
		-ldflags "${LD_FLAGS}" \
		-o ${BUILD_DIR}/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/service/main.go

install: install-go-modules install-linter

.PHONY: install-linter
install-linter:
ifneq "$(INSTALLED_LINT_VERSION)" "$(LATEST_LINT_VERSION)"
	@echo "new golangci-lint version found:" $(LATEST_LINT_VERSION)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin latest
endif

.PHONY: install-go-modules
install-go-modules:
	go mod vendor

# run static analysis tools, configuration in ./.golangci.yml file
.PHONY: lint
lint: install-linter
	golangci-lint run ./...

.PHONY: vulncheck
vulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	./bin/govulncheck-with-excludes.sh ./...

.PHONY: test
test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: test-integration
test-integration:
	docker compose down && docker compose run --build --rm integration-test && docker compose down

.PHONY: test-load
test-load:
	docker compose down && docker compose run --build --rm load-test && docker compose down
