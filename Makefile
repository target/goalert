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
PG_VERSION=13

# add all files except those under web/src/build and web/src/cypress
NODE_DEPS=.gitrev $(shell find web/src -path web/src/build -prune -o -path web/src/cypress -prune -o -type f -print) web/src/app/editor/expr-parser.ts node_modules

# Use sha256sum on linux and shasum -a 256 on mac
SHA_CMD := $(shell if [ -x "$(shell command -v sha256sum 2>/dev/null)" ]; then echo "sha256sum"; else echo "shasum -a 256"; fi)

export CY_ACTION = open
export CY_BROWSER = chrome

export CGO_ENABLED = 0
export PATH := $(PWD)/bin:$(PWD)/bin/tools:$(PATH)
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

CONTAINER_TOOL ?= docker

all: test

release: container-demo container-goalert bin/goalert-linux-amd64.tgz bin/goalert-linux-arm.tgz bin/goalert-linux-arm64.tgz bin/goalert-darwin-amd64.tgz bin/goalert-windows-amd64.zip ## Build all release artifacts

Makefile.binaries.mk: devtools/genmake/*
	go run ./devtools/genmake >$@

$(BIN_DIR)/tools/k6: k6.version
	go run ./devtools/gettool -t k6 -v $(shell cat k6.version) -o $@

$(BIN_DIR)/tools/protoc: protoc.version
	go run ./devtools/gettool -t protoc -v $(shell cat protoc.version) -o $@


$(BIN_DIR)/tools/mailpit: mailpit.version
	go run ./devtools/gettool -t mailpit -v $(shell cat mailpit.version) -o $@

$(BIN_DIR)/tools/sqlc: sqlc.version
	go run ./devtools/gettool -t sqlc -v $(shell cat sqlc.version) -o $@

$(BIN_DIR)/tools/bun: bun.version
	go run ./devtools/gettool -t bun -v $(shell cat bun.version) -o $@

bun.lock: $(BIN_DIR)/tools/bun
	$(BIN_DIR)/tools/bun install

node_modules: $(BIN_DIR)/tools/bun package.json bun.lock
	$(BIN_DIR)/tools/bun install
	touch "$@"

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

cypress: bin/goalert.cover bin/psql-lite bin/pgmocktime $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/build/integration/cypress/plugins/index.js node_modules
	$(BIN_DIR)/tools/bun run cypress install

cy-wide: cypress ## Start cypress tests in desktop mode with dev build in UI mode
	GOALERT_VERSION=$(GIT_VERSION) CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 go run ./devtools/runproc -f Procfile.cypress
cy-mobile: cypress ## Start cypress tests in mobile mode with dev build in UI mode
	GOALERT_VERSION=$(GIT_VERSION) CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 go run ./devtools/runproc -f Procfile.cypress
cy-wide-prod: web/src/build/static/app.js cypress ## Start cypress tests in desktop mode with production build in UI mode
	GOALERT_VERSION=$(GIT_VERSION) CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 CY_ACTION=$(CY_ACTION) go run ./devtools/runproc -f $(PROD_CY_PROC)
cy-mobile-prod: web/src/build/static/app.js cypress ## Start cypress tests in mobile mode with production build in UI mode
	GOALERT_VERSION=$(GIT_VERSION) CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 CY_ACTION=$(CY_ACTION) go run ./devtools/runproc -f $(PROD_CY_PROC)
cy-wide-prod-run: web/src/build/static/app.js cypress ## Start cypress tests in desktop mode with production build in headless mode
	rm -rf test/coverage/integration/cypress-wide
	mkdir -p test/coverage/integration/cypress-wide
	GOCOVERDIR=test/coverage/integration/cypress-wide $(MAKE) $(MFLAGS) cy-wide-prod CY_ACTION=run CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) BUNDLE=1 GOALERT_VERSION=$(GIT_VERSION) 
cy-mobile-prod-run: web/src/build/static/app.js cypress ## Start cypress tests in mobile mode with production build in headless mode
	rm -rf test/coverage/integration/cypress-mobile
	mkdir -p test/coverage/integration/cypress-mobile
	GOCOVERDIR=test/coverage/integration/cypress-mobile $(MAKE) $(MFLAGS) cy-mobile-prod CY_ACTION=run CONTAINER_TOOL=$(CONTAINER_TOOL) PG_VERSION=$(PG_VERSION) BUNDLE=1 GOALERT_VERSION=$(GIT_VERSION) 

swo/swodb/queries.sql.go: $(BIN_DIR)/tools/sqlc sqlc.yaml swo/*/*.sql migrate/migrations/*.sql */queries.sql */*/queries.sql migrate/schema.sql
	$(BIN_DIR)/tools/sqlc generate

web/src/schema.d.ts: graphql2/schema.graphql graphql2/graph/*.graphqls web/src/genschema.go
	go generate ./web/src

help: ## Show all valid options
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

start: bin/goalert bin/mockoidc $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/tools/prometheus $(BIN_DIR)/tools/mailpit ## Start the developer version of the application
	@if [ -d ".vscode" ]; then \
		echo "Detected .vscode directory, running 'vscode' target"; \
		$(MAKE) vscode; \
	fi
	go run ./devtools/waitfor -timeout 1s  "$(DB_URL)" || make postgres
	GOALERT_VERSION=$(GIT_VERSION) GOALERT_STRICT_EXPERIMENTAL=1 go run ./devtools/runproc -f Procfile -l Procfile.local

start-prod: web/src/build/static/app.js bin/mockoidc $(BIN_DIR)/tools/prometheus $(BIN_DIR)/tools/mailpit ## Start the production version of the application
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	$(MAKE) $(MFLAGS) bin/goalert BUNDLE=1
	go run ./devtools/runproc -f Procfile.prod -l Procfile.local

reset-swo: bin/psql-lite bin/goalert bin/waitfor
	./bin/waitfor -timeout 1s  "$(DB_URL)" || make postgres
	./bin/goalert migrate --db-url=postgres://goalert@localhost/goalert
	./bin/psql-lite -d postgres://goalert@localhost -c "update switchover_state set current_state = 'idle'; truncate table switchover_log; drop database if exists goalert2; create database goalert2;"
	./bin/goalert migrate --db-url=postgres://goalert@localhost/goalert2
start-swo: reset-swo bin/goalert bin/mockoidc bin/runproc $(NODE_DEPS) web/src/schema.d.ts $(BIN_DIR)/tools/prometheus $(BIN_DIR)/tools/mailpit ## Start the developer version of the application in switchover mode (SWO)
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

start-integration: web/src/build/static/app.js bin/goalert bin/psql-lite bin/waitfor bin/runproc bin/procwrap $(BIN_DIR)/tools/prometheus $(BIN_DIR)/tools/mailpit reset-integration
	GOALERT_DB_URL="$(INT_DB_URL)" ./bin/runproc -f Procfile.integration

jest: $(NODE_DEPS)
	$(BIN_DIR)/tools/bun run jest $(JEST_ARGS)

test: $(NODE_DEPS) jest $(BIN_DIR)/tools/mailpit ## Run all unit tests
	rm -rf $(PWD)/test/coverage/unit
	mkdir -p $(PWD)/test/coverage/unit
	go test -coverpkg=./... -short ./... -args -test.gocoverdir=$(PWD)/test/coverage/unit

check: check-go check-js ## Run all lint checks
	./devtools/ci/tasks/scripts/codecheck.sh

check-js: generate $(NODE_DEPS)
	$(BIN_DIR)/tools/bun run fmt
	$(BIN_DIR)/tools/bun run lint
	$(BIN_DIR)/tools/bun run check

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
	$(BIN_DIR)/tools/bun run build-storybook --test --quiet 2>/dev/null
	$(BIN_DIR)/tools/bun run playwright install chromium
	$(BIN_DIR)/tools/bun run concurrently -k -s first -n "SB,TEST" -c "magenta,blue" \
		"$(BIN_DIR)/tools/bun run http-server storybook-static -a 127.0.0.1 --port 6008 --silent" \
		"./bin/waitfor tcp://localhost:6008 && $(BIN_DIR)/tools/bun run test-storybook --ci --url http://127.0.0.1:6008 --maxWorkers 2"

storybook: $(NODE_DEPS) # Start the Storybook UI
	$(BIN_DIR)/tools/bun run storybook

playwright-run: $(NODE_DEPS) bin/mockoidc web/src/build/static/app.js bin/goalert.cover web/src/schema.d.ts $(BIN_DIR)/tools/prometheus $(BIN_DIR)/tools/mailpit reset-integration ## Start playwright tests in headless mode
	rm -rf test/coverage/integration/playwright
	mkdir -p test/coverage/integration/playwright
	$(BIN_DIR)/tools/bun run playwright install chromium
	GOCOVERDIR=test/coverage/integration/playwright $(BIN_DIR)/tools/bun run playwright test

playwright-ui: $(NODE_DEPS) bin/mockoidc web/src/build/static/app.js bin/goalert web/src/schema.d.ts $(BIN_DIR)/tools/prometheus reset-integration $(BIN_DIR)/tools/mailpit ## Start the Playwright UI
	$(BIN_DIR)/tools/bun run playwright install chromium
	$(BIN_DIR)/tools/bun run playwright test --ui

smoketest: $(BIN_DIR)/tools/mailpit
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
	go get -u golang.org/x/tools/cmd/stringer




web/src/app/editor/expr-parser.ts: web/src/app/editor/expr.grammar node_modules
	# we need to use .tmp.ts as the extension because lezer-generator will append .ts to the output file
	bin/tools/bun run lezer-generator $< --noTerms --typeScript -o $@.tmp.ts
	bin/tools/bun run prettier -l --write $@.tmp.ts
	cat $@.tmp.ts | sed "s/You probably shouldn't edit it./DO NOT EDIT/" >$@
	rm $@.tmp.ts

web/src/build/static/explore.js: web/src/build/static/app.js
web/src/build/static/app.js: $(NODE_DEPS)
	rm -rf web/src/build/static
	mkdir -p web/src/build/static
	cp -f web/src/app/public/icons/favicon-* web/src/app/public/logos/lightmode_* web/src/app/public/logos/darkmode_* web/src/build/static/
	# used for email templates
	cp web/src/app/public/logos/goalert-alt-logo.png web/src/build/static/
	GOALERT_VERSION=$(GIT_VERSION) $(BIN_DIR)/tools/bun run esbuild --prod
	touch "$@"

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
	rm -rf bin node_modules web/src/node_modules .pnp.cjs .pnp.loader.mjs web/src/build/static .yarn/cache .yarn/install-state.gz .yarn/unplugged storybook-static

new-migration:
	@test "$(NAME)" != "" || (echo "NAME is required" && false)
	@test ! -f migrate/migrations/*-$(NAME).sql || (echo "Migration already exists with the name $(NAME)." && false)
	@echo "-- +migrate Up\n\n\n-- +migrate Down\n" >migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql
	@echo "Created: migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql"

vscode:
	echo "make vscode is no longer necessary since the migration to bun"

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
