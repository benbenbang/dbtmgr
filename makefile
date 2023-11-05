.EXPORT_ALL_VARIABLES:
NAME = statectl
GOARCH = amd64
GOARCH-ARM = arm64
PKG = statectl
Version = $(or $(shell cat version.info), "v0.1.0")
BuildTime = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BuildCommit = $(shell git rev-parse HEAD)
CORES = $(or $(shell sysctl -n hw.ncpu), 1)

ifneq (,$$("ls ./.env"))
include .env
export $(shell sed 's/=.*//' .env)
DBT_STATE_BUCKET := $(DBT_STATE_BUCKET)
DBT_STATE_KEY := $(DBT_STATE_KEY)
DBT_LOCK_KEY := $(DBT_LOCK_KEY)
endif

.PHONY: build
## Build for all platforms
build: clean build-darwin build-linux build-windows

.PHONY: build-cache
## Build for all platforms with cache
build-cache: clean-build build-darwin build-linux build-windows

.PHONY: build-darwin
## Build for MacOS amd64
build-darwin:
	@for BINARY_NAME in ${NAME} ${KUBEPLUGIN}; do \
		for ARCH in ${GOARCH} ${GOARCH-ARM}; do \
			mkdir -p ./build; \
			GOOS=darwin GOARCH=$${ARCH} go build -p ${CORES} -v -o ./build/$${BINARY_NAME}-darwin-$${ARCH} \
			-ldflags="-s -w \
                -X ${PKG}/internal/config.DBT_STATE_BUCKET=${DBT_STATE_BUCKET} \
                -X ${PKG}/internal/config.DBT_STATE_KEY=${DBT_STATE_KEY} \
				-X ${PKG}/internal/config.DBT_LOCK_KEY=${DBT_LOCK_KEY} \
				-X ${PKG}/internal/config.Version=${Version} " \
			main.go; \
		done; \
	done


.PHONY: build-linux
## Build for Linux amd64
build-linux:
	@for BINARY_NAME in ${NAME} ${KUBEPLUGIN}; do \
		for ARCH in ${GOARCH} ${GOARCH-ARM}; do \
			mkdir -p ./build; \
			GOOS=linux GOARCH=$${ARCH} go build -p ${CORES} -v -o ./build/$${BINARY_NAME}-linux-$${ARCH} \
			-ldflags="-s -w \
                -X ${PKG}/internal/config.DBT_STATE_BUCKET=${DBT_STATE_BUCKET} \
                -X ${PKG}/internal/config.DBT_STATE_KEY=${DBT_STATE_KEY} \
				-X ${PKG}/internal/config.DBT_LOCK_KEY=${DBT_LOCK_KEY} \
				-X ${PKG}/internal/config.Version=${Version} " \
			main.go; \
		done; \
	done

.PHONY: build-windows
## Build for Windows
build-windows:
	@for BINARY_NAME in ${NAME} ${KUBEPLUGIN}; do \
		for ARCH in ${GOARCH} ${GOARCH-ARM}; do \
			mkdir -p ./build; \
			GOOS=windows GOARCH=$${ARCH} go build -p ${CORES} -v -o ./build/$${BINARY_NAME}-windows-$${ARCH}.exe \
			-ldflags="-s -w \
                -X ${PKG}/internal/config.DBT_STATE_BUCKET=${DBT_STATE_BUCKET} \
                -X ${PKG}/internal/config.DBT_STATE_KEY=${DBT_STATE_KEY} \
				-X ${PKG}/internal/config.DBT_LOCK_KEY=${DBT_LOCK_KEY} \
				-X ${PKG}/internal/config.Version=${Version} " \
			main.go; \
		done; \
	done

.PHONY: lint
## Run the built-in Go linter with a strict min confidence
lint:
	golint -min_confidence=0.6 ./...

.PHONY: test
## Run all the tests (unit & integration)
test: test-unit test-integration

.PHONY: test-unit
## Run the unit tests
test-unit:
	go tool cover -func=coverage.out && \
	go test -tags=integration ./...

.PHONY: test-integration
## Run the integration tests
test-integration:
	GOOS=linux GOARCH=${GOARCH} go build -i -v -o ${NAME}-linux-${GOARCH} -ldflags="-s -w" ${PKG} && \
	./${NAME}-linux-${GOARCH}

.PHONY: clean
## Remove build files
clean:
	@rm -r ./build 2>/dev/null || true
	@rm coverage.html 2>/dev/null || true
	@rm coverage.out 2>/dev/null || true
	@go clean -cache
	@go clean -testcache

.PHONY: clean-build
clean-build:
	@rm -r ./build 2>/dev/null || true

.DEFAULT_GOAL := help

help:
	@echo "$$(tput bold)Available rules:$$(tput sgr0)"
	@echo
	@sed -n -e "/^## / { \
		h; \
		s/.*//; \
		:doc" \
		-e "H; \
		n; \
		s/^## //; \
		t doc" \
		-e "s/:.*//; \
		G; \
		s/\\n## /---/; \
		s/\\n/ /g; \
		p; \
	}" ${MAKEFILE_LIST} \
	| LC_ALL='C' sort --ignore-case \
	| awk -F '---' \
		-v ncol=$$(tput cols) \
		-v indent=19 \
		-v col_on="$$(tput setaf 6)" \
		-v col_off="$$(tput sgr0)" \
	'{ \
		printf "%s%*s%s ", col_on, -indent, $$1, col_off; \
		n = split($$2, words, " "); \
		line_length = ncol - indent; \
		for (i = 1; i <= n; i++) { \
			line_length -= length(words[i]) + 1; \
			if (line_length <= 0) { \
				line_length = ncol - indent - length(words[i]) - 1; \
				printf "\n%*s ", -indent, " "; \
			} \
			printf "%s ", words[i]; \
		} \
		printf "\n"; \
	}' \
	| more $(shell test $(shell uname) = Darwin && echo '--no-init --raw-control-chars')
