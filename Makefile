.PHONY: stop start build-docker lint tools regendb resetdb
.PHONY: smoketest generate check all test test-long install install-race
.PHONY: cy-wide cy-mobile cy-wide-prod cy-mobile-prod cypress postgres
.PHONY: config.json.bak jest new-migration check-all cy-wide-prod-run cy-mobile-prod-run
.SUFFIXES:

GOFILES = $(shell find . -path ./web/src -prune -o -path ./vendor -prune -o -path ./.git -prune -o -type f -name "*.go" -print | grep -v web/inline_data_gen.go) go.sum
INLINER = devtools/inliner/*.go
CFGPARAMS = devtools/configparams/*.go
DB_URL = postgres://goalert@localhost:5432/goalert?sslmode=disable

LOG_DIR=
GOPATH=$(shell go env GOPATH)
BIN_DIR=bin

GIT_COMMIT=$(shell git rev-parse HEAD || echo '?')
GIT_TREE=$(shell git diff-index --quiet HEAD -- && echo clean || echo dirty)
GIT_VERSION=$(shell git describe --tags --dirty --match 'v*' || echo dev-$(shell date -u +"%Y%m%d%H%M%S"))
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

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
	GOFILES += web/inline_data_gen.go
endif

all: test install

$(BIN_DIR)/runjson: go.sum devtools/runjson/*.go
	go build -o $@ ./devtools/$(@F)
$(BIN_DIR)/psql-lite: go.sum devtools/psql-lite/*.go
	go build -o $@ ./devtools/$(@F)
$(BIN_DIR)/waitfor: go.sum devtools/waitfor/*.go
	go build -o $@ ./devtools/$(@F)
$(BIN_DIR)/simpleproxy: go.sum devtools/simpleproxy/*.go
	go build -o $@ ./devtools/$(@F)
$(BIN_DIR)/mockslack: go.sum $(shell find ./devtools/mockslack -name '*.go')
	go build -o $@ ./devtools/mockslack/cmd/mockslack
$(BIN_DIR)/goalert: go.sum $(GOFILES) graphql2/mapconfig.go
	go build -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert

install: $(GOFILES)
	go install -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" ./cmd/goalert

cypress: bin/runjson bin/waitfor bin/simpleproxy bin/mockslack bin/goalert bin/psql-lite web/src/node_modules
	web/src/node_modules/.bin/cypress install

cy-wide: cypress web/src/build/vendorPackages.dll.js
	CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 bin/runjson $(RUNJSON_ARGS) <devtools/runjson/localdev-cypress.json
cy-mobile: cypress web/src/build/vendorPackages.dll.js
	CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 bin/runjson $(RUNJSON_ARGS) <devtools/runjson/localdev-cypress.json
cy-wide-prod: web/inline_data_gen.go cypress
	CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 CY_ACTION=$(CY_ACTION) bin/runjson $(RUNJSON_ARGS) <$(RUNJSON_PROD_FILE)
cy-mobile-prod: web/inline_data_gen.go cypress
	CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 CY_ACTION=$(CY_ACTION) bin/runjson $(RUNJSON_ARGS) <$(RUNJSON_PROD_FILE)
cy-wide-prod-run: web/inline_data_gen.go cypress
	make cy-wide-prod CY_ACTION=run
cy-mobile-prod-run: web/inline_data_gen.go cypress
	make cy-mobile-prod CY_ACTION=run

start: bin/waitfor web/src/node_modules web/src/build/vendorPackages.dll.js bin/runjson
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	make bin/goalert BUILD_TAGS+=sql_highlight
	bin/runjson <devtools/runjson/localdev.json

jest: web/src/node_modules
	cd web/src && node_modules/.bin/jest $(JEST_ARGS)

test: web/src/node_modules jest
	go test -short ./...

check: generate web/src/node_modules
	# go run devtools/ordermigrations/main.go -check
	go vet ./...
	go run github.com/gordonklaus/ineffassign .
	CGO_ENABLED=0 go run honnef.co/go/tools/cmd/staticcheck ./...
	(cd web/src && yarn fmt)
	./devtools/ci/tasks/scripts/codecheck.sh

check-all: check test smoketest cy-wide-prod-run cy-mobile-prod-run

migrate/inline_data_gen.go: migrate/migrations migrate/migrations/*.sql $(INLINER)
	go generate ./migrate

graphql2/mapconfig.go: $(CFGPARAMS) config/config.go
	(cd ./graphql2 && go run ../devtools/configparams/main.go -out mapconfig.go && goimports -w ./mapconfig.go) || go generate ./graphql2

graphql2/generated.go: graphql2/schema.graphql graphql2/gqlgen.yml
	go generate ./graphql2

generate:
	go generate ./...

smoketest: install bin/goalert
	(cd smoketest && go test -parallel 10 -timeout 20m)

test-migrations: migrate/inline_data_gen.go bin/goalert
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

web/src/yarn.lock: web/src/package.json
	(cd web/src && yarn --no-progress --silent && touch yarn.lock)

web/src/node_modules: web/src/node_modules/.bin/cypress
	touch web/src/node_modules

web/src/node_modules/.bin/cypress: web/src/yarn.lock
	(cd web/src && yarn --no-progress --silent --frozen-lockfile && touch node_modules/.bin/cypress)

web/src/build/index.html: web/src/webpack.prod.config.js web/src/yarn.lock $(shell find ./web/src/app -type f )
	rm -rf web/src/build/static
	(cd web/src && yarn --no-progress --silent --frozen-lockfile && node_modules/.bin/webpack --config webpack.prod.config.js)

web/inline_data_gen.go: web/src/build/index.html $(CFGPARAMS) $(INLINER)
	go generate ./web

web/src/build/vendorPackages.dll.js: web/src/node_modules web/src/webpack.dll.config.js
	(cd web/src && node_modules/.bin/webpack --config ./webpack.dll.config.js --progress)

config.json.bak: bin/goalert
	bin/goalert get-config "--db-url=$(DB_URL)" 2>/dev/null >config.json.new
	(test -s config.json.new && test "`cat config.json.new`" != "{}" && mv config.json.new config.json.bak || rm -f config.json.new)

postgres:
	docker run -d \
		--restart=always \
		-e POSTGRES_USER=goalert \
		--name goalert-postgres \
		-p 5432:5432 \
		postgres:11-alpine || docker start goalert-postgres

regendb: bin/goalert migrate/inline_data_gen.go config.json.bak
	go run ./devtools/resetdb --with-rand-data
	test -f config.json.bak && bin/goalert set-config --allow-empty-data-encryption-key "--db-url=$(DB_URL)" <config.json.bak || true
	bin/goalert add-user --admin --email admin@example.com --user admin --pass admin123 "--db-url=$(DB_URL)"

resetdb: migrate/inline_data_gen.go config.json.bak
	go run ./devtools/resetdb --no-migrate

clean:
	git clean -xdf ./web ./bin ./vendor ./smoketest

build-docker: bin/goalert bin/mockslack

lint: $(GOFILES)
	go run github.com/golang/lint/golint $(shell go list ./...)

new-migration:
	@test "$(NAME)" != "" || (echo "NAME is required" && false)
	@test ! -f migrate/migrations/*-$(NAME).sql || (echo "Migration already exists with the name $(NAME)." && false)
	@echo "-- +migrate Up\n\n\n-- +migrate Down\n" >migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql
	@echo "Created: migrate/migrations/$(shell date +%Y%m%d%H%M%S)-$(NAME).sql"
