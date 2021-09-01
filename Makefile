.PHONY: stop start build-docker lint tools regendb resetdb
.PHONY: smoketest generate check all test test-long install install-race
.PHONY: cy-wide cy-mobile cy-wide-prod cy-mobile-prod cypress postgres
.PHONY: config.json.bak jest new-migration check-all cy-wide-prod-run cy-mobile-prod-run
.PHONY: docker-goalert docker-all-in-one release force-yarn
.SUFFIXES:

include Makefile.binaries.mk

CFGPARAMS = devtools/configparams/*.go
DB_URL = postgres://goalert@localhost:5432/goalert?sslmode=disable

LOG_DIR=
GOPATH:=$(shell go env GOPATH)

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

DOCKER_IMAGE_PREFIX=docker.io/goalert
DOCKER_TAG=$(GIT_VERSION)

ifeq ($(PUSH), 1)
PUSH_FLAG=--push
endif


all: test install

release: docker-goalert docker-all-in-one bin/goalert-linux-amd64.tgz bin/goalert-linux-arm.tgz bin/goalert-linux-arm64.tgz bin/goalert-darwin-amd64.tgz
docker-all-in-one: bin/linux-amd64/goalert bin/linux-arm64/goalert bin/linux-arm/goalert bin/linux-amd64/resetdb bin/linux-arm64/resetdb bin/linux-arm/resetdb
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/all-in-one-demo:$(DOCKER_TAG) -f devtools/ci/dockerfiles/all-in-one/Dockerfile.buildx .
docker-goalert: bin/linux-amd64/goalert bin/linux-arm64/goalert bin/linux-arm/goalert
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/goalert:$(DOCKER_TAG) -f devtools/ci/dockerfiles/goalert/Dockerfile.buildx .

Makefile.binaries.mk: devtools/genmake/*
	go run ./devtools/genmake >$@

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
	bin/waitfor -timeout 1s  "$(DB_URL)" || make postgres
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
	# go run ./devtools/ordermigrations -check
	go vet ./...
	go run github.com/gordonklaus/ineffassign ./...
	CGO_ENABLED=0 go run honnef.co/go/tools/cmd/staticcheck ./...
	yarn run fmt
	yarn run lint
	yarn workspaces run check
	./devtools/ci/tasks/scripts/codecheck.sh

check-all: check test smoketest cy-wide-prod-run cy-mobile-prod-run

graphql2/mapconfig.go: $(CFGPARAMS) config/config.go graphql2/generated.go devtools/configparams/*
	(cd ./graphql2 && go run ../devtools/configparams -out mapconfig.go && go run golang.org/x/tools/cmd/goimports -w ./mapconfig.go) || go generate ./graphql2

graphql2/maplimit.go: $(CFGPARAMS) limit/id.go graphql2/generated.go devtools/limitapigen/*
	(cd ./graphql2 && go run ../devtools/limitapigen -out maplimit.go && go run golang.org/x/tools/cmd/goimports -w ./maplimit.go) || go generate ./graphql2

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

notification/desttype_string.go: notification/desttype.go
	go generate ./notification
notification/type_string.go: notice/notice.go
	go generate ./notice

config.json.bak: bin/goalert
	bin/goalert get-config "--db-url=$(DB_URL)" 2>/dev/null >config.json.new || rm config.json.new
	(test -s config.json.new && test "`cat config.json.new`" != "{}" && mv config.json.new config.json.bak || rm -f config.json.new)

postgres: bin/waitfor
	(docker run -d \
		--restart=always \
		-e POSTGRES_USER=goalert \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		--name goalert-postgres \
		-p 5432:5432 \
		postgres:13-alpine && ./bin/waitfor "$(DB_URL)" && make regendb) || docker start goalert-postgres

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
