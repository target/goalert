// +build tools

package devtools

import (
	_ "github.com/gordonklaus/ineffassign"
	_ "github.com/mailhog/MailHog"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/tools/cmd/stringer"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
