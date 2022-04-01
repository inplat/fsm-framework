//go:build tools
// +build tools

//go:generate bash -c "go build -ldflags \"-X 'main.version=$(go list -m -f '{{.Version}}' github.com/golangci/golangci-lint)' -X 'main.commit=test' -X 'main.date=test'\" -o ../bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint"
//go:generate go install github.com/vektra/mockery/v2@latest

// Package tools contains go:generate commands for all project tools with versions stored in local go.mod file
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
