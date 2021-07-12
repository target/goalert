.PHONY: stop start build-docker lint tools regendb resetdb
.PHONY: smoketest generate check all test test-long install install-race
.PHONY: cy-wide cy-mobile cy-wide-prod cy-mobile-prod cypress postgres
.PHONY: config.json.bak jest new-migration check-all cy-wide-prod-run cy-mobile-prod-run
.PHONY: docker-goalert docker-all-in-one release force-yarn
.SUFFIXES:

GOALERT_DEPS := $(shell find . -path ./web/src -prune -o -path ./vendor -prune -o -path ./.git -prune -o -type f -name "*.go" -print) go.sum
CFGPARAMS = devtools/configparams/*.go
DB_URL = postgres://goalert@localhost:5432/goalert?sslmode=disable

LOG_DIR=
GOPATH:=$(shell go env GOPATH)
BIN_DIR=bin

GIT_COMMIT:=$(shell git rev-parse HEAD || echo '?')
GIT_TREE:=$(shell git diff-index --quiet HEAD -- && echo clean || echo dirty)
GIT_VERSION:=$(shell git describe --tags --dirty --match 'v*' || echo dev-$(shell date -u +"%Y%m%d%H%M%S"))
BUILD_DATE:=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=

export ZONEINFO=$(shell go env GOROOT)/lib/time/zoneinfo.zip

LD_FLAGS+=-X github.com/target/goalert/version.gitCommit=$(GIT_COMMIT)
LD_FLAGS+=-X github.com/target/goalert/version.gitVersion=$(GIT_VERSION)
LD_FLAGS+=-X github.com/target/goalert/version.gitTreeState=$(GIT_TREE)
LD_FLAGS+=-X github.com/target/goalert/version.buildDate=$(BUILD_DATE)

export CY_ACTION = open
export CY_BROWSER = chrome
export RUNJSON_PROD_FILE = devtools/runjson/localdev-cypress-prod.json

ifdef LOG_DIR
RUNJSON_ARGS += -logs=$(LOG_DIR)
endif

export CGO_ENABLED = 0
export PATH := $(PWD)/bin:$(PATH)
export GOOS = $(shell go env GOOS)
export GOALERT_DB_URL_NEXT = $(DB_URL_NEXT)

ifeq ($(shell test -d vendor && echo -n yes), yes)
export GOFLAGS=-mod=vendor
endif

ifdef BUNDLE
	GOALERT_DEPS += web/src/build/static/app.js
endif

DOCKER_IMAGE_PREFIX=docker.io/goalert
DOCKER_TAG=$(GIT_VERSION)

ifeq ($(PUSH), 1)
PUSH_FLAG=--push
endif

GOALERT_DEPS += migrate/migrations/ migrate/migrations/*.sql graphql2/graphqlapp/playground.html web/index.html
GOALERT_DEPS += graphql2/mapconfig.go graphql2/maplimit.go graphql2/generated.go graphql2/models_gen.go

all: test install

release: docker-goalert docker-all-in-one bin/goalert-linux-amd64.tgz bin/goalert-linux-arm.tgz bin/goalert-linux-arm64.tgz bin/goalert-darwin-amd64.tgz
docker-all-in-one: bin/goalert-linux-amd64 bin/goalert-linux-arm bin/goalert-linux-arm64 bin/resetdb-linux-amd64 bin/resetdb-linux-arm bin/resetdb-linux-arm64
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/all-in-one-demo:$(DOCKER_TAG) -f devtools/ci/dockerfiles/all-in-one/Dockerfile.buildx .
docker-goalert: bin/goalert-linux-amd64 bin/goalert-linux-arm bin/goalert-linux-arm64
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/goalert:$(DOCKER_TAG) -f devtools/ci/dockerfiles/goalert/Dockerfile.buildx .

$(BIN_DIR)/goalert: go.sum $(GOALERT_DEPS) graphql2/mapconfig.go
	go build $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-linux-amd64: $(BIN_DIR)/goalert web/src/build/static/app.js
	GOOS=linux go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-smoketest-linux-amd64: $(BIN_DIR)/goalert
	GOOS=linux go test ./smoketest -c -o $@
$(BIN_DIR)/goalert-linux-arm: $(BIN_DIR)/goalert web/src/build/static/app.js
	GOOS=linux GOARCH=arm GOARM=7 go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-linux-arm64: $(BIN_DIR)/goalert web/src/build/static/app.js
	GOOS=linux GOARCH=arm64 go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-darwin-amd64: $(BIN_DIR)/goalert web/src/build/static/app.js
	GOOS=darwin go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert

$(BIN_DIR)/%-linux-amd64: go.mod go.sum $(shell find ./devtools -type f) migrate/*.go migrate/migrations/ migrate/migrations/*.sql
	GOOS=linux go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)
$(BIN_DIR)/%-linux-arm: go.mod go.sum $(shell find ./devtools -type f) migrate/*.go migrate/migrations/ migrate/migrations/*.sql
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)
$(BIN_DIR)/%-linux-arm64: go.mod go.sum $(shell find ./devtools -type f) migrate/*.go migrate/migrations/ migrate/migrations/*.sql
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)

$(BIN_DIR)/goalert-%.tgz: $(BIN_DIR)/goalert-%
	rm -rf $(BIN_DIR)/$*
	mkdir -p $(BIN_DIR)/$*/goalert/bin
	cp $(BIN_DIR)/goalert-$* $(BIN_DIR)/$*/goalert/bin/goalert
	tar czvf $@ -C $(BIN_DIR)/$* goalert

$(BIN_DIR)/%: go.mod go.sum $(shell find ./devtools -type f) migrate/*.go migrate/migrations/ migrate/migrations/*.sql
	go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)

$(BIN_DIR)/integration/goalert/cypress.json: web/src/cypress.json
	sed 's/\.ts/\.js/' web/src/cypress.json >bin/integration/goalert/cypress.json

$(BIN_DIR)/integration/goalert/cypress: node_modules web/src/webpack.cypress.js $(BIN_DIR)/integration/goalert/cypress.json $(shell find ./web/src/cypress)
	rm -rf $@
	yarn workspace goalert-web webpack --config webpack.cypress.js --target node
	cp -r web/src/cypress/fixtures bin/integration/goalert/cypress/
	touch $@

$(BIN_DIR)/integration/goalert/bin: $(BIN_DIR)/goalert-linux-amd64 $(BIN_DIR)/goalert-smoketest-linux-amd64 $(BIN_DIR)/mockslack-linux-amd64 $(BIN_DIR)/simpleproxy-linux-amd64 $(BIN_DIR)/pgdump-lite-linux-amd64 $(BIN_DIR)/waitfor-linux-amd64 $(BIN_DIR)/runjson-linux-amd64 $(BIN_DIR)/psql-lite-linux-amd64 $(BIN_DIR)/procwrap-linux-amd64
	rm -rf $@
	mkdir -p bin/integration/goalert/bin
	cp bin/*-linux-amd64 bin/integration/goalert/bin/
	for f in bin/integration/goalert/bin/*-linux-amd64; do ln -s $$(basename $$f) bin/integration/goalert/bin/$$(basename $$f -linux-amd64); done
	touch $@

$(BIN_DIR)/integration/goalert/devtools: $(shell find ./devtools/ci)
	rm -rf $@
	mkdir -p bin/integration/goalert/devtools
	cp -r devtools/ci bin/integration/goalert/devtools/
	touch $@

$(BIN_DIR)/integration/goalert/.git: $(shell find ./.git)
	rm -rf $@
	mkdir -p bin/integration/goalert/.git
	git rev-parse HEAD >bin/integration/goalert/COMMIT
	test -d .git/resource && cp -r .git/resource bin/integration/goalert/.git/ || true

$(BIN_DIR)/integration: $(BIN_DIR)/integration/goalert/.git $(BIN_DIR)/integration/goalert/devtools $(BIN_DIR)/integration/goalert/cypress $(BIN_DIR)/integration/goalert/bin
	touch $@

$(BIN_DIR)/integration.tgz: bin/integration
	tar czvf bin/integration.tgz -C bin/integration goalert

$(BIN_DIR)/tools/protoc: protoc.version
	go run ./devtools/gettool -t protoc -v $(shell cat protoc.version) -o $@

$(BIN_DIR)/tools/prometheus: prometheus.version
	go run ./devtools/gettool -t prometheus -v $(shell cat prometheus.version) -o $@

$(BIN_DIR)/tools/protoc-gen-go: go.mod
	GOBIN=$(abspath $(BIN_DIR))/tools go install google.golang.org/protobuf/cmd/protoc-gen-go
$(BIN_DIR)/tools/protoc-gen-go-grpc: go.mod
	GOBIN=$(abspath $(BIN_DIR))/tools go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

system.ca.pem:
	go run ./cmd/goalert gen-cert ca
system.ca.key:
	go run ./cmd/goalert gen-cert ca
plugin.ca.pem:
	go run ./cmd/goalert gen-cert ca
plugin.ca.key:
	go run ./cmd/goalert gen-cert ca

goalert-server.pem: system.ca.pem system.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert server
goalert-server.key: system.ca.pem system.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert server
goalert-server.ca.pem: system.ca.pem system.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert server

goalert-client.pem: system.ca.pem plugin.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert client
goalert-client.key: system.ca.pem plugin.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert client
goalert-client.ca.pem: system.ca.pem plugin.ca.key plugin.ca.pem
	go run ./cmd/goalert gen-cert client

install: $(GOALERT_DEPS)
	go install $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" ./cmd/goalert

cypress: bin/runjson bin/waitfor bin/procwrap bin/simpleproxy bin/mockslack bin/goalert bin/psql-lite node_modules web/src/schema.d.ts
	yarn cypress install

cy-wide: cypress
	CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 bin/runjson $(RUNJSON_ARGS) <devtools/runjson/localdev-cypress.json
cy-mobile: cypress
	CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 bin/runjson $(RUNJSON_ARGS) <devtools/runjson/localdev-cypress.json
cy-wide-prod: web/src/build/static/app.js cypress
	CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 CY_ACTION=$(CY_ACTION) bin/runjson $(RUNJSON_ARGS) <$(RUNJSON_PROD_FILE)
cy-mobile-prod: web/src/build/static/app.js cypress
	CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 CY_ACTION=$(CY_ACTION) bin/runjson $(RUNJSON_ARGS) <$(RUNJSON_PROD_FILE)
cy-wide-prod-run: web/src/build/static/app.js cypress
	make cy-wide-prod CY_ACTION=run
cy-mobile-prod-run: web/src/build/static/app.js cypress
	make cy-mobile-prod CY_ACTION=run

web/src/schema.d.ts: graphql2/schema.graphql node_modules web/src/genschema.go devtools/gqlgen/*
	go generate ./web/src

start: bin/waitfor node_modules bin/runjson web/src/schema.d.ts $(BIN_DIR)/tools/prometheus
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	make bin/goalert BUILD_TAGS+=sql_highlight
	GOALERT_VERSION=$(GIT_VERSION) bin/runjson <devtools/runjson/localdev.json

start-prod: bin/waitfor web/src/build/static/app.js bin/runjson $(BIN_DIR)/tools/prometheus
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	make bin/goalert BUILD_TAGS+=sql_highlight BUNDLE=1
	bin/runjson <devtools/runjson/localdev-prod.json

jest: node_modules 
	yarn workspace goalert-web run jest $(JEST_ARGS)

test: node_modules jest
	go test -short ./...

force-yarn:
	yarn install --no-progress --silent --frozen-lockfile --check-files

check: force-yarn generate node_modules
	# go run devtools/ordermigrations/main.go -check
	go vet ./...
	go run github.com/gordonklaus/ineffassign ./...
	CGO_ENABLED=0 go run honnef.co/go/tools/cmd/staticcheck ./...
	yarn run fmt
	yarn run lint
	yarn workspaces run check
	./devtools/ci/tasks/scripts/codecheck.sh

check-all: check test smoketest cy-wide-prod-run cy-mobile-prod-run

graphql2/mapconfig.go: $(CFGPARAMS) config/config.go graphql2/generated.go devtools/configparams/main.go
	(cd ./graphql2 && go run ../devtools/configparams/main.go -out mapconfig.go && go run golang.org/x/tools/cmd/goimports -w ./mapconfig.go) || go generate ./graphql2

graphql2/maplimit.go: $(CFGPARAMS) limit/id.go graphql2/generated.go devtools/limitapigen/main.go
	(cd ./graphql2 && go run ../devtools/limitapigen/main.go -out maplimit.go && go run golang.org/x/tools/cmd/goimports -w ./maplimit.go) || go generate ./graphql2

graphql2/generated.go: graphql2/schema.graphql graphql2/gqlgen.yml go.mod
	go generate ./graphql2

pkg/sysapi/sysapi_grpc.pb.go: pkg/sysapi/sysapi.proto $(BIN_DIR)/tools/protoc-gen-go-grpc $(BIN_DIR)/tools/protoc
	PATH="$(BIN_DIR)/tools" protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/sysapi/sysapi.proto
pkg/sysapi/sysapi.pb.go: pkg/sysapi/sysapi.proto $(BIN_DIR)/tools/protoc-gen-go $(BIN_DIR)/tools/protoc
	PATH="$(BIN_DIR)/tools" protoc --go_out=. --go_opt=paths=source_relative pkg/sysapi/sysapi.proto

generate: node_modules pkg/sysapi/sysapi.pb.go
	go generate ./...

smoketest:
	(cd smoketest && go test -parallel 10 -timeout 20m)

test-migrations: bin/goalert
	(cd smoketest && go test -run TestMigrations)

tools:
	go get -u golang.org/x/tools/cmd/gorename
	go get -u golang.org/x/tools/cmd/present
	go get -u golang.org/x/tools/cmd/bundle
	go get -u golang.org/x/tools/cmd/gomvpkg
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gordonklaus/ineffassign
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u golang.org/x/tools/cmd/stringer

yarn.lock: package.json web/src/package.json Makefile
	yarn --no-progress --silent --check-files && touch $@

node_modules/.yarn-integrity: yarn.lock Makefile
	yarn install --no-progress --silent --frozen-lockfile --check-files
	touch $@

node_modules: yarn.lock node_modules/.yarn-integrity
	touch -c $@

web/src/build/static/app.js: web/src/webpack.prod.config.js node_modules $(shell find ./web/src/app -type f ) web/src/schema.d.ts
	rm -rf web/src/build/static
	yarn workspace goalert-web webpack --config webpack.prod.config.js --env=GOALERT_VERSION=$(GIT_VERSION)

notification/desttype_string.go: notification/dest.go
	go generate ./notification
notification/type_string.go: notice/notice.go
	go generate ./notice

config.json.bak: bin/goalert
	bin/goalert get-config "--db-url=$(DB_URL)" 2>/dev/null >config.json.new || rm config.json.new
	(test -s config.json.new && test "`cat config.json.new`" != "{}" && mv config.json.new config.json.bak || rm -f config.json.new)

postgres:
	docker run -d \
		--restart=always \
		-e POSTGRES_USER=goalert \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		--name goalert-postgres \
		-p 5432:5432 \
		postgres:13-alpine || docker start goalert-postgres

regendb: bin/resetdb bin/goalert config.json.bak
	./bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000001
	test -f config.json.bak && bin/goalert set-config --allow-empty-data-encryption-key "--db-url=$(DB_URL)" <config.json.bak || true
	bin/goalert add-user --user-id=00000000-0000-0000-0000-000000000001 --user admin --pass admin123 "--db-url=$(DB_URL)"

resetdb: config.json.bak
	go run ./devtools/resetdb --no-migrate

clean:
	git clean -xdf

build-docker: bin/goalert bin/mockslack

lint: $(GOALERT_DEPS)
	go run github.com/golang/lint/golint $(shell go list ./...)

new-migration:
	@test "$(NAME)" != "" || (echo "NAME is required" && false)
	@test ! -f migrate/migrations/*-$(NAME).sql || (echo "Migration already exists with the name $(NAME)." && false)
	@echo "-- +migrate Up\n\n\n-- +migrate Down\n" >migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql
	@echo "Created: migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql"
