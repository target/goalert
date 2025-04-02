//go:build tools
// +build tools

package devtools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/kffl/speedbump"
	_ "github.com/riverqueue/river/cmd/river"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/tools/cmd/stringer"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
