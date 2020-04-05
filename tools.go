// +build tools

// This file lists all go binary dependencies used by build scripts so that they are managed properly by go mod
package main

import (
	_ "github.com/cespare/reflex"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "gotest.tools/gotestsum"
)
