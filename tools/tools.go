package tools

// Put only installable tools into this list.
// scripts/install_build_tools.sh parses these imports to install them.
import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
