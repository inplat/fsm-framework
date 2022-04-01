TIMEOUT := 10s
FSM_GENERATOR = fsm-generator
LOCATION = $(shell pwd)
FLAGS = -tags 'osusergo netgo static_build'
OS = $(shell uname -nsm)
BUILD_COMMIT = $(shell git log --format="%H" -n 1)
LDFLAGS = "-X 'main.BuildTime=$(NOW)' -X 'main.BuildOSUname=$(OS)' -X 'main.BuildCommit=$(BUILD_COMMIT)'

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

GOSRC  = $(shell find . -type f -name '*.go' | grep -v /tools/)

# Устанавливает бинарки для тулзов: linter, etc.
.PHONY: tools
tools:
	cd tools && go mod tidy && go generate -tags tools

# Линт
.PHONY: lint
lint:
	$(info #Running lint...)
	./bin/golangci-lint run --config=.golangci.yml ./...

# Линт на отличия от мастера (без скачивания тулзов)
.PHONY: fast-lint
fast-lint:
	$(info #Running lint...)
	./bin/golangci-lint run --new-from-rev=origin/master --config=.golangci.yml ./...

# Запускает юнит тесты
.PHONY: test
test:
	go test -timeout $(TIMEOUT) ./...

# Генерация всего
.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build:
	@go version
	go build -o bin/$(FSM_GENERATOR) $(FLAGS) -ldflags $(LDFLAGS) $(LOCATION)/cmd/$(FSM_GENERATOR)