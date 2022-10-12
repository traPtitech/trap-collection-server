//go:build tools
// +build tools

package main

import (
	_ "github.com/cosmtrek/air"
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/google/wire/cmd/wire"
)
