#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
VERSION ?= $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/evmos/os.git
DOCKER := $(shell which docker)

export GO111MODULE = on

# Default target executed when no arguments are given to make.
default_target: all

.PHONY: default_target

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit
test-all: test-unit test-race

# For unit tests we don't want to execute the upgrade tests in tests/e2e but
# we want to include all unit tests in the subfolders (tests/e2e/*)
PACKAGES_UNIT=$(shell go list ./... | grep -v '/tests/e2e$$')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-timeout=15m
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)

test-race: ARGS=-race
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

test-unit-cover: ARGS=-timeout=15m -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

test-scripts:
	@echo "Running scripts tests"
	@pytest -s -vv ./scripts

test-solidity:
	@echo "Beginning solidity tests..."
	./scripts/run-solidity-tests.sh

.PHONY: run-tests test test-all $(TEST_TARGETS)

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)

.PHONY: benchmark

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab
	solhint contracts/**/*.sol

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0

lint-fix-contracts:
	@cd contracts && \
	npm i && \
	npm run lint-fix
	solhint --fix contracts/**/*.sol

.PHONY: lint lint-fix

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.pb.gw.go' | xargs gofumpt -w -l

.PHONY: format


format-python: format-isort format-black

format-black:
	find . -name '*.py' -type f -not -path "*/node_modules/*" | xargs black

format-isort:
	find . -name '*.py' -type f -not -path "*/node_modules/*" | xargs isort

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.11.6
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace --user 0 $(protoImageName)

protoLintVer=0.44.0
protoLinterImage=yoheimuta/protolint
protoLinter=$(DOCKER) run --rm -v "$(CURDIR):/workspace" --workdir /workspace --user 0 $(protoLinterImage):$(protoLintVer)

# ------
# NOTE: If you are experiencing problems running these commands, try deleting
#       the docker images and execute the desired command again.
#
proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "generating implementations from Protobuf files"
	@$(protoImage) sh ./scripts/generate_protos.sh

proto-format:
	@echo "formatting Protobuf files"
	@$(protoImage) find ./ -name *.proto -exec clang-format -i {} \;

proto-lint:
	@echo "linting Protobuf files"
	@$(protoImage) buf lint --error-format=json
	@$(protoLinter) lint ./proto

proto-check-breaking:
	@echo "checking Protobuf files for breaking changes"
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking

###############################################################################
###                                Releasing                                ###
###############################################################################

PACKAGE_NAME:=github.com/evmos/os
GOLANG_CROSS_VERSION  = v1.22
GOPATH ?= '$(HOME)/go'
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v ${GOPATH}/pkg:/go/pkg \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip validate --skip publish --snapshot

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip validate

.PHONY: release-dry-run release

###############################################################################
###                        Compile Solidity Contracts                       ###
###############################################################################

# Install the necessary dependencies, compile the solidity contracts found in the
# evmOS repository and then clean up the contracts data.
contracts-all: contracts-compile contracts-clean

# Clean smart contract compilation artifacts, dependencies and cache files
contracts-clean:
	@echo "Cleaning up the contracts directory..."
	@python3 ./scripts/compile_smart_contracts/compile_smart_contracts.py --clean

# Compile the solidity contracts found in the evmOS repository.
contracts-compile:
	@echo "Compiling smart contracts..."
	@python3 ./scripts/compile_smart_contracts/compile_smart_contracts.py --compile

# Add a new solidity contract to be compiled
contracts-add:
	@echo "Adding a new smart contract to be compiled..."
	@python3 ./scripts/compile_smart_contracts/compile_smart_contracts.py --add $(CONTRACT)

###############################################################################
###                           Miscellaneous Checks                          ###
###############################################################################

# TODO: turn into CI action
check-licenses:
	@echo "Checking licenses..."
	@curl -sSfL https://raw.githubusercontent.com/evmos/evmos/v19.0.0/scripts/license_checker/check_licenses.py -o check_licenses.py
	@python3 check_licenses.py .
	@rm check_licenses.py

# Use changelog linter and auto-fix from https://github.com/MalteHerrmann/changelog-utils
cluVersion=v1.1.2
cluImage=ghcr.io/malteherrmann/changelog-utils
clu=$(DOCKER) run --rm -v "$(CURDIR):/workspace" --workdir /workspace --user 0 $(cluImage):$(cluVersion)

changelog-check:
	@echo "Checking changelog..."
	@$(clu) lint

changelog-fix:
	@echo "Fixing changelog..."
	@$(clu) fix
