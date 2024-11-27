.PHONY:{{range $.Builds}} $(BIN_DIR)/{{.Name}}/_all{{end}}
.PHONY:{{range $.ContainerArch}} container-goalert-{{.}} container-demo-{{.}}{{end}} container-goalert container-demo container-goalert-manifest container-demo-manifest

BIN_DIR=bin
GO_DEPS := Makefile.binaries.mk $(shell find . -path ./web/src -prune -o -path ./vendor -prune -o -path ./.git -prune -o -type f -name "*.go" -print) go.sum
GO_DEPS += migrate/migrations/ migrate/migrations/*.sql web/index.html graphql2/graphqlapp/slack.manifest.yaml swo/*/*.sql
GO_DEPS += graphql2/mapconfig.go graphql2/maplimit.go graphql2/generated.go graphql2/models_gen.go
GO_DEPS += web/explore.html web/live.js .gitrev

ifdef BUNDLE
	GO_DEPS += web/src/build/static/app.js web/src/build/static/explore.js
endif

GIT_COMMIT:=$(shell git rev-parse HEAD || echo '?')
GIT_TREE:=$(shell git diff-index --quiet HEAD -- && echo clean || echo dirty)
GIT_VERSION:=$(shell git describe --tags --dirty --match 'v*' || echo dev)
BUILD_DATE:=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=

# update .gitrev with GIT_VERSION (unless it matches the current value)
ifneq ($(shell cat .gitrev 2>/dev/null),$(GIT_VERSION))
$(shell echo "$(GIT_VERSION)" > .gitrev)
endif
.gitrev:
	@echo "$(GIT_VERSION)" > .gitrev

export ZONEINFO:=$(shell go env GOROOT)/lib/time/zoneinfo.zip

LD_FLAGS+=-X github.com/target/goalert/version.gitCommit=$(GIT_COMMIT)
LD_FLAGS+=-X github.com/target/goalert/version.gitVersion=$(GIT_VERSION)
LD_FLAGS+=-X github.com/target/goalert/version.gitTreeState=$(GIT_TREE)
LD_FLAGS+=-X github.com/target/goalert/version.buildDate=$(BUILD_DATE)

IMAGE_REPO=docker.io/goalert
IMAGE_TAG=$(GIT_VERSION)

PUSH:=0

PUSH_ARG=
ifeq ($(PUSH),1)
	PUSH_ARG=--push
endif

PREBUILT:=.prebuilt

container-goalert:{{range $.ContainerArch}} bin/goalert-linux-{{.}}.tgz{{end}}
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm -t $(IMAGE_REPO)/goalert:$(IMAGE_TAG) -f devtools/ci/dockerfiles/goalert/Dockerfile$(PREBUILT) $(PUSH_ARG) .

container-demo:{{range $.ContainerArch}} bin/goalert-linux-{{.}}.tgz bin/linux-{{.}}/resetdb{{end}}
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm -t $(IMAGE_REPO)/demo:$(IMAGE_TAG) -f devtools/ci/dockerfiles/demo/Dockerfile$(PREBUILT) $(PUSH_ARG) .

$(BIN_DIR)/build/integration/cypress/plugins/index.js: package.json yarn.lock web/src/esbuild.cypress.js $(shell find ./web/src/cypress)
	rm -rf $@
	yarn run esbuild-cy
	mkdir -p $@/plugins
	cp web/src/cypress/plugins/index.js $@/plugins/index.js
	touch $@

{{range $.Builds}}
$(BIN_DIR)/build/integration/bin/build/goalert-{{.Name}}: $(BIN_DIR)/build/goalert-{{.Name}}
	rm -rf $@
	mkdir -p $@
	cp -r $(BIN_DIR)/build/goalert-{{.Name}}/goalert $@/
	touch $@
{{end}}

$(BIN_DIR)/build/integration/devtools: $(shell find ./devtools/ci)
	rm -rf $@
	mkdir -p $@
	cp -r devtools/ci $@/
	touch $@

$(BIN_DIR)/build/integration/.git: $(shell find ./.git)
	rm -rf $@
	mkdir -p $@
	test -d .git/resource && cp -r .git/resource $@/ || true
	touch $@

$(BIN_DIR)/build/integration/COMMIT: $(BIN_DIR)/build/integration/.git
	git rev-parse HEAD >$@

$(BIN_DIR)/build/integration: $(BIN_DIR)/build/integration/.git $(BIN_DIR)/build/integration/COMMIT $(BIN_DIR)/build/integration/devtools $(BIN_DIR)/build/integration/cypress {{- range $.Builds}} $(BIN_DIR)/build/integration/bin/build/goalert-{{.Name}}{{end}}
	touch $@

{{range $tool := $.Tools}}
{{if eq $tool.Name "goalert"}}
$(BIN_DIR)/{{$tool.Name}}.cover: $(GO_DEPS) {{$tool.Deps}}
	go build {{$tool.Flags}} -cover -coverpkg=./... -o $@ ./{{$tool.Dir}}
{{end}}
$(BIN_DIR)/{{$tool.Name}}: $(GO_DEPS) {{$tool.Deps}}
	go build {{$tool.Flags}} -o $@ ./{{$tool.Dir}}
{{range $build := $.Builds}}
$(BIN_DIR)/{{$build.Name}}/{{$tool.Name}}{{$build.Ext}}: $(GO_DEPS) {{$tool.Deps}} {{$tool.ProdDeps}}
	{{$build.Env}} go build -trimpath {{$tool.Flags}} -o $@ ./{{$tool.Dir}}
{{end}}
{{end}}

{{range $build := $.Builds}}
$(BIN_DIR)/{{$build.Name}}/_all: $(BIN_DIR)/{{$build.Name}}/goalert-smoketest{{range $tool := $.Tools}} $(BIN_DIR)/{{$build.Name}}/{{$tool.Name}}{{$build.Ext}}{{end}}

$(BIN_DIR)/{{$build.Name}}/goalert-smoketest: $(GO_DEPS)
	{{$build.Env}} go test ./smoketest -c -o $@
{{end}}
$(BIN_DIR)/goalert-smoketest: $(GO_DEPS)
	go test ./smoketest -c -o $@

{{range $bundle := $.Bundles}}
{{range $build := $.Builds}}
$(BIN_DIR)/build/{{$bundle.Name}}-{{$build.Name}}:{{range $name := $bundle.Binaries}} $(BIN_DIR)/{{$build.Name}}/{{$name}}{{$build.Ext}}{{end}}{{range $bundle.CopyDir}} {{.}}{{end}}
	rm -rf $@
	mkdir -p $@/{{$bundle.DirName}}/bin/
	cp {{range $name := $bundle.Binaries}} $(BIN_DIR)/{{$build.Name}}/{{$name}}{{$build.Ext}}{{end}} $@/{{$bundle.DirName}}/bin/
	{{- if $bundle.CopyDir}}
	cp -r {{range $bundle.CopyDir}} {{.}}/.{{end}} $@/{{$bundle.DirName}}/
	{{- end}}
	touch $@

$(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}.tgz: $(BIN_DIR)/build/{{$bundle.Name}}-{{$build.Name}}
	tar -czvf $(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}.tgz -C $(BIN_DIR)/build/{{$bundle.Name}}-{{$build.Name}}/ .

$(BIN_DIR)/{{$bundle.Name}}-{{$build.Name}}.zip: $(BIN_DIR)/build/{{$bundle.Name}}-{{$build.Name}}
	rm -f $@
	cd $(BIN_DIR)/build/{{$bundle.Name}}-{{$build.Name}} && zip -r $(abspath $@) .
{{end}}
{{end}}
