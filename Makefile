#
# Copyright 2019 New Relic Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Name of binary to create.
BINARY = newrelic-istio-adapter

# Semver of application.
VERSION := $(shell cat ./VERSION)

# Go related variables.
GOPATH ?= $(shell pwd)
GOBIN ?= $(GOPATH)/bin
GO = GOPATH=$(GOPATH) GOBIN=$(GOBIN) go

# While statically linking we want to inject version related information into the binary.
LDFLAGS = -ldflags="-extldflags '-static' -X main.Version=$(VERSION)"


all: generate test build

generate:
# The generate script requires local copies of dependencies.
	@$(GO) mod vendor
	@$(GO) generate ./...

TAGS ?= ""
TEST_TAGS := --tags=$(TAGS)
test:
	@$(GO) test -v $(TEST_TAGS) ./...

build:
	@$(GO) build $(LDFLAGS) -o $(GOBIN)/$(BINARY) cmd/main.go

.PHONY: generate test build
