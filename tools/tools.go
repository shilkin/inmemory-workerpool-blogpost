//go:build tools

package tools

//go:generate go install github.com/daixiang0/gci
//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install golang.org/x/tools/cmd/goimports
//go:generate go install mvdan.cc/gofumpt

import (
	_ "github.com/daixiang0/gci"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "mvdan.cc/gofumpt"
)
