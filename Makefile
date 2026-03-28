PROJECT_NAME := ccc

.PHONY: build test test-go test-shell test-installers test-version ci

build:
	go build ./cmd/ccc

test-go:
	go test ./...

test-version:
	bash tests/check_version.sh

test-shell:
	bash tests/test.sh

test-installers:
	bash tests/install_requires_node.sh
	bash tests/install_best_effort.sh

test: test-version test-go test-shell test-installers

ci: test build
