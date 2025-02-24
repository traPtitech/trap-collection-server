//go:build tools
// +build tools

package main

import (
	_ "github.com/cosmtrek/air"
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "go.uber.org/mock/mockgen"
)
