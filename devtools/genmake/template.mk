.PHONY:{{range $.Builds}} $(BIN_DIR)/{{.Name}}/_all{{end}}

BIN_DIR=bin
GO_DEPS := Makefile.binaries.mk $(shell find . -path ./web/src -prune -o -path ./vendor -prune -o -path ./.git -prune -o -type f -name "*.go" -print) go.sum
GO_DEPS += migrate/migrations/ migrate/migrations/*.sql graphql2/graphqlapp/playground.html web/index.html graphql2/graphqlapp/slack.manifest.yaml
GO_DEPS += graphql2/mapconfig.go graphql2/maplimit.go graphql2/generated.go graphql2/models_gen.go

ifdef BUNDLE
	GO_DEPS += web/src/build/static/app.js
endif

GIT_COMMIT:=$(shell git rev-parse HEAD || echo '?')
GIT_TREE:=$(shell git diff-index --quiet HEAD -- && echo clean || echo dirty)
GIT_VERSION:=$(shell git describe --tags --dirty --match 'v*' || echo dev-$(shell date -u +"%Y%m%d%H%M%S"))
BUILD_DATE:=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=

export ZONEINFO:=$(shell go env GOROOT)/lib/time/zoneinfo.zip

LD_FLAGS+=-X github.com/target/goalert/version.gitCommit=$(GIT_COMMIT)
LD_FLAGS+=-X github.com/target/goalert/version.gitVersion=$(GIT_VERSION)
LD_FLAGS+=-X github.com/target/goalert/version.gitTreeState=$(GIT_TREE)
LD_FLAGS+=-X github.com/target/goalert/version.buildDate=$(BUILD_DATE)

{{range $tool := $.Tools}}
$(BIN_DIR)/{{$tool.Name}}: $(GO_DEPS) {{$tool.Deps}}
	go build -o $@ ./{{$tool.Dir}}
{{range $build := $.Builds}}
$(BIN_DIR)/{{$build.Name}}/{{$tool.Name}}: $(GO_DEPS) {{$tool.Deps}} {{$tool.ProdDeps}}
	{{$build.Env}} go build -trimpath {{$tool.Flags}} -o $@ ./{{$tool.Dir}}
{{end}}
{{end}}

{{range $build := $.Builds}}
$(BIN_DIR)/{{$build.Name}}/_all: $(BIN_DIR)/{{$build.Name}}/goalert-smoketest{{range $tool := $.Tools}} $(BIN_DIR)/{{$build.Name}}/{{$tool.Name}}{{end}}

$(BIN_DIR)/{{$build.Name}}/goalert-smoketest: $(GO_DEPS)
	{{$build.Env}} go test ./smoketest -c -o $@
{{end}}
$(BIN_DIR)/goalert-smoketest: $(GO_DEPS)
	go test ./smoketest -c -o $@

{{range $bundle := $.Bundles}}
{{range $build := $.Builds}}
$(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}.tgz:{{range $name := $bundle.Binaries}} $(BIN_DIR)/{{$build.Name}}/{{$name}}{{end}}{{range $bundle.Copy}} {{.}}{{end}}
	rm -rf $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}/
	mkdir -p $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}/{{$bundle.DirName}}/bin/
	cp {{range $name := $bundle.Binaries}} $(BIN_DIR)/{{$build.Name}}/{{$name}}{{end}} $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}/{{$bundle.DirName}}/bin/
	{{- if $bundle.Copy}}
	cp -r {{range $bundle.Copy}} {{.}}{{end}} $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}/{{$bundle.DirName}}/
	{{- end}}
	tar -czvf $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}.tgz -C $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}/ .
{{end}}
{{end}}
