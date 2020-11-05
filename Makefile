# SPDX-License-Identifier: BSD-3-Clause
#
# Authors: Alexander Jung <alex@nderjung.net>
#
# Copyright (c) 2020, Alexander Jung.  All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
#
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the names of its
#    contributors may be used to endorse or promote products derived from
#    this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

BIN        = limp
BUILD_DIR  = ./dist/limp_linux_amd64
BUILD_PATH = $(BUILD_DIR)/$(BIN)
PWD        := ${CURDIR}
REGISTRY   = docker.io
TEST_IMAGE = busybox:latest

.PHONY: all
all: clean build

## For CI

.PHONY: ci-unit-test
ci-unit-test:
	go test -cover -v -race ./...

.PHONY: ci-static-analysis
ci-static-analysis:
	go vet ./...
	gofmt -s -l . 2>&1 | grep -vE '^\.git/' | grep -vE '^\.cache/'
	golangci-lint run

.PHONY: ci-install-go-tools
ci-install-go-tools:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sudo sh -s -- -b /usr/local/bin/ latest

.PHONY: ci-install-ci-tools
ci-install-ci-tools:
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sudo sh -s -- -b /usr/local/bin/ "v0.146.0"

.PHONY: ci-docker-login
ci-docker-login:
	echo '${DOCKER_PASSWORD}' | docker login -u '${DOCKER_USERNAME}' --password-stdin '${REGISTRY}'

.PHONY: ci-docker-logout
ci-docker-logout:
	docker logout '${REGISTRY}'

.PHONY: ci-publish-release
ci-publish-release:
	goreleaser --rm-dist

.PHONY: ci-build-snapshot-packages
ci-build-snapshot-packages:
	goreleaser \
		--snapshot \
		--skip-publish \
		--rm-dist

.PHONY: ci-release
ci-release:
	goreleaser release --rm-dist

.PHONY: ci-test-production-image
ci-test-production-image:
	docker run --rm -t \
		${REGISTRY}/ndrjng/limp:latest \
			--version

.PHONY: ci-test-deb-package-install
ci-test-deb-package-install:
	docker run \
		-v //var/run/docker.sock://var/run/docker.sock \
		-v /${PWD}://src \
		-w //src \
		ubuntu:latest \
			/bin/bash -x -c "\
				apt update && \
				apt install ./dist/limp_*_linux_amd64.deb -y && \
				limp --version \
			"

ci-test-rpm-package-install:
	docker run \
		-v //var/run/docker.sock://var/run/docker.sock \
		-v /${PWD}://src \
		-w //src \
		fedora:latest \
			/bin/bash -x -c "\
				dnf install ./dist/limp_*_linux_amd64.rpm -y && \
				limp --version \
			"

.PHONY: ci-test-linux-run
ci-test-linux-run:
	chmod 755 ./dist/limp_linux_amd64/limp && \
	./dist/limp_linux_amd64/limp --version

.PHONY: ci-test-mac-run
ci-test-mac-run:
	chmod 755 ./dist/limp_darwin_amd64/limp && \
	./dist/limp_darwin_amd64/limp --version

#
# For development
#

.PHONY: run
run: build
	$(BUILD_PATH) build -t limp-example:latest -f .data/Dockerfile.example .

.PHONY: build
build:
	go build -o $(BUILD_PATH)

.PHONY: devenv
devenv:
	docker run -it --rm \
		-v $(PWD):/go/src/github.com/nderjung/limp \
		-w /go/src/github.com/nderjung/limp \
		golang:1.13 bash

.PHONY: clean
clean:
	rm -rf dist
	go clean


