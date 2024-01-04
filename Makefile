.PHONY: help start tools regendb resetdb db-schema postgres-reset
.PHONY: smoketest generate check all test check-js check-go
.PHONY: cy-wide cy-mobile cy-wide-prod cy-mobile-prod cypress postgres
.PHONY: config.json.bak jest new-migration cy-wide-prod-run cy-mobile-prod-run
.PHONY: goalert-container demo-container release reset-integration yarn ensure-yarn vscode upgrade-js playwright-ui
.PHONY: timezone/zones.txt timezone/aliases.txt
.SUFFIXES:

default: bin/goalert

include Makefile.binaries.mk

CFGPARAMS = devtools/configparams/*.go
DB_URL = postgres://goalert@localhost:5432/goalert
INT_DB = goalert_integration
INT_DB_URL = $(shell go run ./devtools/scripts/db-url "$(DB_URL)" "$(INT_DB)")

SWO_DB_MAIN = goalert_swo_main
SWO_DB_NEXT = goalert_swo_next
SWO_DB_URL_MAIN = $(shell go run ./devtools/scripts/db-url "$(DB_URL)" "$(SWO_DB_MAIN)")
SWO_DB_URL_NEXT = $(shell go run ./devtools/scripts/db-url "$(DB_URL)" "$(SWO_DB_NEXT)")

LOG_DIR=
GOPATH:=$(shell go env GOPATH)
YARN_VERSION=3.6.3
PG_VERSION=13

NODE_DEPS=.pnp.cjs .yarnrc.yml

# Use sha256sum on linux and shasum -a 256 on mac
SHA_CMD := $(shell if [ -x "$(shell command -v sha256sum 2>/dev/null)" ]; then echo "sha256sum"; else echo "shasum -a 256"; fi)

export CY_ACTION = open
export CY_BROWSER = chrome

export CGO_ENABLED = 0
export PATH := $(PWD)/bin:$(PATH)
export GOOS = $(shell go env GOOS)
export GOALERT_DB_URL_NEXT = $(DB_URL_NEXT)

PROD_CY_PROC = Procfile.cypress.prod
SIZE:=1

PUBLIC_URL := http://localhost:3030$(HTTP_PREFIX)
export GOALERT_PUBLIC_URL := $(PUBLIC_URL)

# used to enable experimental features, use `goalert --list-experimental` to see available features or check the expflag package
EXPERIMENTAL :=
export GOALERT_EXPERIMENTAL := $(EXPERIMENTAL)

ifeq ($(CI), 1)
PROD_CY_PROC = Procfile.cypress.ci
endif


ifeq ($(PUSH), 1)
PUSH_FLAG=--push
endif

all: test

release: container-demo container-goalert bin/goalert-linux-amd64.tgz bin/goalert-linux-arm.tgz bin/goalert-linux-arm64.tgz bin/goalert-darwin-amd64.tgz bin/goalert-windows-amd64.zip ## Build all release artifacts

Makefile.binaries.mk: devtools/genmake/*
	go run ./devtools/genmake >$@

$(BIN_DIR)/tools/protoc: protoc.version
	go run ./devtools/gettool -t protoc -v $(shell cat protoc.version) -o $@

$(BIN_DIR)/tools/sqlc: sqlc.version
	go run ./devtools/gettool -t sqlc -v $(shell cat sqlc.version) -o $@

$(BIN_DIR)/tools/prometheus: prometheus.version
	go run ./devtools/gettool -t prometheus -v $(shell cat prometheus.version) -o $@

$(BIN_DIR)/tools/golangci-lint: golangci-lint.version
	go run ./devtools/gettool -t golangci-lint -v $(shell cat golangci-lint.version) -o $@

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

cypress: bin/goalert.cover bin/psql-lite bin/pgmocktime $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/build/integration/cypress/plugins/index.js
	$(MAKE) ensure-yarn
	yarn cypress install

cy-wide: cypress
	CONTAINER_TOOL=$(CONTAINER_TOOL) CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 go run ./devtools/runproc -f Procfile.cypress
cy-mobile: cypress
	CONTAINER_TOOL=$(CONTAINER_TOOL) CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 go run ./devtools/runproc -f Procfile.cypress
cy-wide-prod: web/src/build/static/app.js cypress
	CONTAINER_TOOL=$(CONTAINER_TOOL) CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 CY_ACTION=$(CY_ACTION) go run ./devtools/runproc -f $(PROD_CY_PROC)
cy-mobile-prod: web/src/build/static/app.js cypress
	CONTAINER_TOOL=$(CONTAINER_TOOL) CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 CY_ACTION=$(CY_ACTION) go run ./devtools/runproc -f $(PROD_CY_PROC)
cy-wide-prod-run: web/src/build/static/app.js cypress
	rm -rf test/coverage/integration/cypress-wide
	mkdir -p test/coverage/integration/cypress-wide
	GOCOVERDIR=test/coverage/integration/cypress-wide $(MAKE) $(MFLAGS) cy-wide-prod CY_ACTION=run CONTAINER_TOOL=$(CONTAINER_TOOL) BUNDLE=1
cy-mobile-prod-run: web/src/build/static/app.js cypress
	rm -rf test/coverage/integration/cypress-mobile
	mkdir -p test/coverage/integration/cypress-mobile
	GOCOVERDIR=test/coverage/integration/cypress-mobile $(MAKE) $(MFLAGS) cy-mobile-prod CY_ACTION=run CONTAINER_TOOL=$(CONTAINER_TOOL) BUNDLE=1

swo/swodb/queries.sql.go: $(BIN_DIR)/tools/sqlc sqlc.yaml swo/*/*.sql migrate/migrations/*.sql */queries.sql */*/queries.sql migrate/schema.sql
	$(BIN_DIR)/tools/sqlc generate

web/src/schema.d.ts: graphql2/schema.graphql graphql2/graph/*.graphqls $(NODE_DEPS) web/src/genschema.go
	go generate ./web/src

help: ## Show all valid options
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

start: bin/goalert bin/mockoidc $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/tools/prometheus ## Start the developer version of the application
	@if [ -d ".vscode" ]; then \
		echo "Detected .vscode directory, running 'vscode' target"; \
		$(MAKE) vscode; \
	fi
	go run ./devtools/waitfor -timeout 1s  "$(DB_URL)" || make postgres
	GOALERT_VERSION=$(GIT_VERSION) GOALERT_STRICT_EXPERIMENTAL=1 go run ./devtools/runproc -f Procfile -l Procfile.local

start-prod: web/src/build/static/app.js bin/mockoidc $(BIN_DIR)/tools/prometheus ## Start the production version of the application
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	$(MAKE) $(MFLAGS) bin/goalert BUNDLE=1
	go run ./devtools/runproc -f Procfile.prod -l Procfile.local


start-swo: bin/psql-lite bin/goalert bin/waitfor bin/mockoidc bin/runproc $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/tools/prometheus ## Start the developer version of the application in switchover mode (SWO)
	./bin/waitfor -timeout 1s  "$(DB_URL)" || make postgres
	./bin/goalert migrate --db-url=postgres://goalert@localhost/goalert
	./bin/psql-lite -d postgres://goalert@localhost -c "update switchover_state set current_state = 'idle'; truncate table switchover_log; drop database if exists goalert2; create database goalert2;"
	./bin/goalert migrate --db-url=postgres://goalert@localhost/goalert2
	GOALERT_VERSION=$(GIT_VERSION) ./bin/runproc -f Procfile.swo -l Procfile.local

reset-integration: bin/waitfor bin/goalert.cover bin/psql-lite bin/resetdb
	rm -rf test/coverage/integration/reset
	mkdir -p test/coverage/integration/reset
	./bin/waitfor -timeout 1s  "$(DB_URL)" || make postgres
	./bin/psql-lite -d "$(DB_URL)" -c 'DROP DATABASE IF EXISTS $(INT_DB); CREATE DATABASE $(INT_DB);'
	./bin/psql-lite -d "$(DB_URL)" -c 'DROP DATABASE IF EXISTS $(SWO_DB_MAIN); CREATE DATABASE $(SWO_DB_MAIN);'
	./bin/psql-lite -d "$(DB_URL)" -c 'DROP DATABASE IF EXISTS $(SWO_DB_NEXT); CREATE DATABASE $(SWO_DB_NEXT);'
	./bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000001 -db-url "$(SWO_DB_URL_MAIN)" -admin-db-url "$(DB_URL)" -mult 0.1
	./bin/goalert.cover add-user --user-id=00000000-0000-0000-0000-000000000001 --user admin --pass admin123 "--db-url=$(SWO_DB_URL_MAIN)"
	GOCOVERDIR=test/coverage/integration/reset ./bin/goalert.cover --db-url "$(INT_DB_URL)" migrate
	./bin/psql-lite -d "$(INT_DB_URL)" -c "insert into users (id, role, name) values ('00000000-0000-0000-0000-000000000001', 'admin', 'Admin McIntegrationFace'),('00000000-0000-0000-0000-000000000002', 'user', 'User McIntegrationFace');"
	GOCOVERDIR=test/coverage/integration/reset ./bin/goalert.cover add-user --db-url "$(INT_DB_URL)" --user-id=00000000-0000-0000-0000-000000000001 --user admin --pass admin123
	GOCOVERDIR=test/coverage/integration/reset ./bin/goalert.cover add-user --db-url "$(INT_DB_URL)" --user-id=00000000-0000-0000-0000-000000000002 --user user --pass user1234
	cat test/integration/setup/goalert-config.json | GOCOVERDIR=test/coverage/integration/reset ./bin/goalert.cover set-config --allow-empty-data-encryption-key --db-url "$(INT_DB_URL)"
	rm -f *.session.json

start-integration: web/src/build/static/app.js bin/goalert bin/psql-lite bin/waitfor bin/runproc bin/procwrap $(BIN_DIR)/tools/prometheus reset-integration
	GOALERT_DB_URL="$(INT_DB_URL)" ./bin/runproc -f Procfile.integration

jest: $(NODE_DEPS)
	$(MAKE) ensure-yarn
	yarn run jest $(JEST_ARGS)

test: $(NODE_DEPS) jest ## Run all unit tests
	rm -rf $(PWD)/test/coverage/unit
	mkdir -p $(PWD)/test/coverage/unit
	go test -coverpkg=./... -short ./... -args -test.gocoverdir=$(PWD)/test/coverage/unit

check: check-go check-js ## Run all lint checks
	./devtools/ci/tasks/scripts/codecheck.sh

.yarnrc.yml: package.json
	$(MAKE) yarn

.yarn/releases/yarn-$(YARN_VERSION).cjs:
	yarn set version $(YARN_VERSION) || $(MAKE) yarn

ensure-yarn: # Yarn ensures the correct version of yarn is installed
	@echo "Checking yarn version..."
	@yarn --version | grep -q -F "$(YARN_VERSION)" || $(MAKE) yarn
	$(MAKE) .yarn/releases/yarn-$(YARN_VERSION).cjs

yarn:
	corepack enable
	corepack prepare yarn@$(YARN_VERSION) --activate

check-js: generate $(NODE_DEPS)
	$(MAKE) ensure-yarn
	yarn install
	yarn run fmt
	yarn run lint
	yarn run check

check-go: generate $(BIN_DIR)/tools/golangci-lint
	@go mod tidy
	# go run ./devtools/ordermigrations -check
	$(BIN_DIR)/tools/golangci-lint run

graphql2/mapconfig.go: $(CFGPARAMS) config/config.go graphql2/generated.go devtools/configparams/*
	(cd ./graphql2 && go run ../devtools/configparams -out mapconfig.go && go run golang.org/x/tools/cmd/goimports -w ./mapconfig.go) || go generate ./graphql2

graphql2/maplimit.go: $(CFGPARAMS) limit/id.go graphql2/generated.go devtools/limitapigen/*
	(cd ./graphql2 && go run ../devtools/limitapigen -out maplimit.go && go run golang.org/x/tools/cmd/goimports -w ./maplimit.go) || go generate ./graphql2

graphql2/generated.go: graphql2/schema.graphql graphql2/gqlgen.yml go.mod graphql2/graph/*.graphqls
	go generate ./graphql2

pkg/sysapi/sysapi_grpc.pb.go: pkg/sysapi/sysapi.proto $(BIN_DIR)/tools/protoc-gen-go-grpc $(BIN_DIR)/tools/protoc
	PATH="$(BIN_DIR)/tools" protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/sysapi/sysapi.proto
pkg/sysapi/sysapi.pb.go: pkg/sysapi/sysapi.proto $(BIN_DIR)/tools/protoc-gen-go $(BIN_DIR)/tools/protoc
	PATH="$(BIN_DIR)/tools" protoc --go_out=. --go_opt=paths=source_relative pkg/sysapi/sysapi.proto

generate: $(NODE_DEPS) pkg/sysapi/sysapi.pb.go pkg/sysapi/sysapi_grpc.pb.go $(BIN_DIR)/tools/sqlc
	$(BIN_DIR)/tools/sqlc generate
	go generate ./...


test-all: test-unit test-components test-smoke test-integration
test-integration: playwright-run cy-wide-prod-run cy-mobile-prod-run
test-smoke: smoketest
test-unit: test

test-components:  $(NODE_DEPS) bin/waitfor
	yarn build-storybook --ci --test --quiet
	yarn concurrently -k -s first -n "SB,TEST" -c "magenta,blue" \
		"yarn http-server storybook-static -a 127.0.0.1 --port 6008 --silent" \
		"./bin/waitfor tcp://localhost:6008 && yarn test-storybook --ci --url http://127.0.0.1:6008"

storybook: $(NODE_DEPS) # Start the Storybook UI
	yarn storybook

bin/MailHog: go.mod go.sum
	go build -o bin/MailHog github.com/mailhog/MailHog

playwright-run: $(NODE_DEPS) bin/mockoidc web/src/build/static/app.js bin/goalert.cover web/src/schema.d.ts $(BIN_DIR)/tools/prometheus reset-integration bin/MailHog
	$(MAKE) ensure-yarn
	rm -rf test/coverage/integration/playwright
	mkdir -p test/coverage/integration/playwright
	yarn playwright install chromium
	GOCOVERDIR=test/coverage/integration/playwright yarn playwright test

playwright-ui: $(NODE_DEPS) bin/mockoidc web/src/build/static/app.js bin/goalert web/src/schema.d.ts $(BIN_DIR)/tools/prometheus reset-integration bin/MailHog ## Start the Playwright UI
	$(MAKE) ensure-yarn
	yarn playwright install chromium
	yarn playwright test --ui

smoketest:
	rm -rf test/coverage/smoke
	mkdir -p test/coverage/smoke
	(cd test/smoke && go test -coverpkg=../../... -parallel 10 -timeout 20m -args -test.gocoverdir=$(PWD)/test/coverage/smoke)

test-migrations: bin/goalert
	(cd test/smoke && go test -run TestMigrations)

db-schema: $(BIN_DIR)/tools/sqlc
	$(BIN_DIR)/tools/sqlc generate -f devtools/pgdump-lite/sqlc.yaml # always run
	go run ./devtools/psql-lite -d "$(DB_URL)" -c 'DROP DATABASE IF EXISTS mk_dump_schema; CREATE DATABASE mk_dump_schema;'
	go run ./migrate/cmd/goalert-migrate --db-url "$(dir $(DB_URL))mk_dump_schema" up
	echo '-- This file is auto-generated by "make db-schema"; DO NOT EDIT' > migrate/schema.sql
	echo "-- DATA=$(shell $(SHA_CMD) migrate/migrations/* | sort | $(SHA_CMD))" >> migrate/schema.sql
	echo "-- DISK=$(shell ls migrate/migrations | sort | $(SHA_CMD))" >> migrate/schema.sql
	echo "-- PSQL=$$(go run ./devtools/psql-lite -d '$(dir $(DB_URL))mk_dump_schema' -c 'select id from gorp_migrations order by id' | sort | $(SHA_CMD))" >> migrate/schema.sql
	go run ./devtools/pgdump-lite/cmd/pgdump-lite -d "$(dir $(DB_URL))mk_dump_schema" -s >> migrate/schema.sql
	go run ./devtools/psql-lite -d "$(DB_URL)" -c 'DROP DATABASE IF EXISTS mk_dump_schema;'

tools:
	go get -u golang.org/x/tools/cmd/gorename
	go get -u golang.org/x/tools/cmd/present
	go get -u golang.org/x/tools/cmd/bundle
	go get -u golang.org/x/tools/cmd/gomvpkg
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gordonklaus/ineffassign
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u golang.org/x/tools/cmd/stringer

.pnp.cjs: yarn.lock Makefile package.json .yarnrc.yml
	$(MAKE) ensure-yarn
	yarn install && touch "$@"


web/src/build/static/explore.js: web/src/build/static

web/src/build/static: web/src/esbuild.config.js $(NODE_DEPS) $(shell find ./web/src/app -type f ) $(shell find ./web/src/explore -type f ) web/src/schema.d.ts
	$(MAKE) ensure-yarn
	rm -rf web/src/build/static
	mkdir -p web/src/build/static
	cp -f web/src/app/public/icons/favicon-* web/src/app/public/logos/lightmode_* web/src/app/public/logos/darkmode_* web/src/build/static/
	GOALERT_VERSION=$(GIT_VERSION) yarn run esbuild

web/src/build/static/app.js: web/src/build/static $(NODE_DEPS)
	

notification/desttype_string.go: notification/desttype.go
	go generate ./notification
notification/type_string.go: notice/notice.go
	go generate ./notice

config.json.bak: bin/goalert
	bin/goalert get-config "--db-url=$(DB_URL)" 2>/dev/null >config.json.new || rm config.json.new
	(test -s config.json.new && test "`cat config.json.new`" != "{}" && mv config.json.new config.json.bak || rm -f config.json.new)

postgres-reset:
	$(CONTAINER_TOOL) rm -f goalert-postgres || true
	$(MAKE) postgres PG_VERSION=$(PG_VERSION)

postgres: bin/waitfor
	($(CONTAINER_TOOL) run -d \
		--restart=always \
		-e POSTGRES_USER=goalert \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		--name goalert-postgres \
		--shm-size 1g \
		-p 5432:5432 \
		docker.io/library/postgres:$(PG_VERSION)-alpine && ./bin/waitfor "$(DB_URL)" && make regendb) || ($(CONTAINER_TOOL) start goalert-postgres && ./bin/waitfor "$(DB_URL)")

regendb: bin/resetdb bin/goalert config.json.bak ## Reset the database and fill it with random data
	./bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000001 -mult $(SIZE)
	test -f config.json.bak && bin/goalert set-config --allow-empty-data-encryption-key "--db-url=$(DB_URL)" <config.json.bak || true
	bin/goalert add-user --user-id=00000000-0000-0000-0000-000000000001 --user admin --pass admin123 "--db-url=$(DB_URL)"

resetdb: config.json.bak ## Recreate the database leaving it empty (no migrations)
	go run ./devtools/resetdb --no-migrate

clean: ## Clean up build artifacts
	rm -rf bin node_modules web/src/node_modules .pnp.cjs .pnp.loader.mjs web/src/build/static .yarn/cache .yarn/install-state.gz .yarn/unplugged

new-migration:
	@test "$(NAME)" != "" || (echo "NAME is required" && false)
	@test ! -f migrate/migrations/*-$(NAME).sql || (echo "Migration already exists with the name $(NAME)." && false)
	@echo "-- +migrate Up\n\n\n-- +migrate Down\n" >migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql
	@echo "Created: migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql"

vscode: $(NODE_DEPS) 
	yarn dlx @yarnpkg/sdks vscode

.yarn/plugins/@yarnpkg/plugin-interactive-tools.cjs: $(NODE_DEPS)
	yarn plugin import interactive-tools

upgrade-js: .yarn/plugins/@yarnpkg/plugin-interactive-tools.cjs ## Interactively upgrade javascript packages
	yarn upgrade-interactive

test/coverage/total.out: test/coverage/integration/*/* test/coverage/*/* Makefile
	rm -rf test/coverage/total
	mkdir -p test/coverage/total
	go tool covdata merge -i test/coverage/integration/cypress-wide,test/coverage/integration/cypress-mobile,test/coverage/integration/playwright,test/coverage/integration/reset,test/coverage/smoke,test/coverage/unit -pcombine -o test/coverage/total

	go tool covdata textfmt -i test/coverage/total -o test/coverage/total.out.tmp
	cat test/coverage/total.out.tmp | grep -v ^github.com/target/goalert/graphql2/generated.go | grep -v ^github.com/target/goalert/graphql2/mapconfig.go | grep -v ^github.com/target/goalert/graphql2/maplimit.go | grep -v ^github.com/target/goalert/pkg/sysapi/sysapi_grpc.pb.go | grep -v ^github.com/target/goalert/pkg/sysapi/sysapi.pb.go | grep -v ^github.com/target/goalert/graphql2/models_gen.go | grep -v ^github.com/target/goalert/gadb | grep -v ^github.com/target/goalert/swo/swodb | grep -v ^github.com/target/goalert/devtools >test/coverage/total.out

test/coverage/report.txt: test/coverage/total.out
	go tool cover -func=test/coverage/total.out | tee test/coverage/report.txt

timezone/zones.txt: # generate a list of all canonical zones
	grep -v '^#' /usr/share/zoneinfo/zone.tab | awk '{print $$3}' >>$@.tmp
# add all non-symlink posix Etc/ zones
	find /usr/share/zoneinfo/Etc -type f ! -type l -exec basename {} \; | awk '{print "Etc/"$$0}' >>$@.tmp
	echo "# This file is auto-generated by 'make timezone/zones.txt'; DO NOT EDIT" >$@
	cat $@.tmp | sort >>$@
	rm $@.tmp

timezone/aliases.txt: # generate a list of ALIAS=ZONE
	find /usr/share/zoneinfo/posix -type l -exec sh -c 'echo "{}=`readlink -e {}`"' \; | sed s@/usr/share/zoneinfo/posix/@@g | grep -v localtime | grep -v posixrules >$@.tmp
	find /usr/share/zoneinfo/posix -maxdepth 1 -type f ! -type l -exec sh -c 'echo "{}={}"' \; | sed s@/usr/share/zoneinfo/posix/@@g >>$@.tmp
	echo "# This file is auto-generated by 'make timezone/aliases.txt'; DO NOT EDIT" >$@
	cat $@.tmp | sort >>$@
	rm $@.tmp
