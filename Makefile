SHELL := /bin/sh

VERSION_FILE ?= VERSION
VERSION ?= $(shell tr -d '[:space:]' < "$(VERSION_FILE)" 2>/dev/null || printf '%s' dev)
DIST_DIR ?= dist

.PHONY: test build install update-local enable-local-updates release release-artifacts smoke-release check-release-version clean

test:
	go test ./...

build:
	go build -ldflags "-X github.com/trues/qbs/internal/cli.Version=$(VERSION)" -o qbs ./cmd/qbs

PREFIX ?= /usr/local

install: build
	mkdir -p "$(PREFIX)/bin"
	cp qbs "$(PREFIX)/bin/qbs"
	chmod 755 "$(PREFIX)/bin/qbs"

update-local:
	PREFIX="$(HOME)/.local" ./scripts/update-local.sh

enable-local-updates:
	git_dir="$$(git rev-parse --git-path hooks)"; mkdir -p "$$git_dir"; cp scripts/post-commit-local "$$git_dir/post-commit"; chmod 755 "$$git_dir/post-commit"

release:
	./scripts/release.sh

release-artifacts:
	VERSION=$(VERSION) DIST_DIR=$(DIST_DIR) ./scripts/build-release.sh

check-release-version:
	./scripts/check-release-version.sh $(VERSION)

smoke-release: release-artifacts
	VERSION=$(VERSION) QBS_RELEASE_DIR=$(DIST_DIR) ./scripts/smoke-release.sh

clean:
	rm -rf $(DIST_DIR) qbs qbs.exe
