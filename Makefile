.PHONY: stop start build-docker lint tools regendb resetdb
.PHONY: smoketest generate check all test test-long install install-race
.PHONY: cy-wide cy-mobile cy-wide-prod cy-mobile-prod cypress postgres
.PHONY: config.json.bak jest new-migration check-all cy-wide-prod-run cy-mobile-prod-run
.PHONY: docker-goalert docker-all-in-one release
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
	GOFILES += web/inline_data_gen.go
endif

DOCKER_IMAGE_PREFIX=docker.io/goalert
DOCKER_TAG=$(GIT_VERSION)

ifeq ($(PUSH), 1)
PUSH_FLAG=--push
endif
GOFILES += graphql2/mapconfig.go graphql2/maplimit.go graphql2/generated.go graphql2/models_gen.go

all: test install

release: docker-goalert docker-all-in-one bin/goalert-linux-amd64.tgz bin/goalert-linux-arm.tgz bin/goalert-linux-arm64.tgz bin/goalert-darwin-amd64.tgz
docker-all-in-one: bin/goalert-linux-amd64 bin/goalert-linux-arm bin/goalert-linux-arm64 bin/resetdb-linux-amd64 bin/resetdb-linux-arm bin/resetdb-linux-arm64
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/all-in-one-demo:$(DOCKER_TAG) -f devtools/ci/dockerfiles/all-in-one/Dockerfile.buildx .
docker-goalert: bin/goalert-linux-amd64 bin/goalert-linux-arm bin/goalert-linux-arm64
	docker buildx build $(PUSH_FLAG) --platform linux/amd64,linux/arm,linux/arm64 -t $(DOCKER_IMAGE_PREFIX)/goalert:$(DOCKER_TAG) -f devtools/ci/dockerfiles/goalert/Dockerfile.buildx .

$(BIN_DIR)/goalert: go.sum $(GOFILES) graphql2/mapconfig.go
	go build $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-linux-amd64: $(BIN_DIR)/goalert web/inline_data_gen.go
	GOOS=linux go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-smoketest-linux-amd64: $(BIN_DIR)/goalert
	GOOS=linux go test ./smoketest -c -o $@
$(BIN_DIR)/goalert-linux-arm: $(BIN_DIR)/goalert web/inline_data_gen.go
	GOOS=linux GOARCH=arm GOARM=7 go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-linux-arm64: $(BIN_DIR)/goalert web/inline_data_gen.go
	GOOS=linux GOARCH=arm64 go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert
$(BIN_DIR)/goalert-darwin-amd64: $(BIN_DIR)/goalert web/inline_data_gen.go
	GOOS=darwin go build -trimpath $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" -o $@ ./cmd/goalert

$(BIN_DIR)/%-linux-amd64: go.mod go.sum $(shell find ./devtools -name '*.go')
	GOOS=linux go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)
$(BIN_DIR)/%-linux-arm: go.mod go.sum $(shell find ./devtools -name '*.go')
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)
$(BIN_DIR)/%-linux-arm64: go.mod go.sum $(shell find ./devtools -name '*.go')
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)

$(BIN_DIR)/goalert-%.tgz: $(BIN_DIR)/goalert-%
	rm -rf $(BIN_DIR)/$*
	mkdir -p $(BIN_DIR)/$*/goalert/bin
	cp $(BIN_DIR)/goalert-$* $(BIN_DIR)/$*/goalert/bin/goalert
	tar czvf $@ -C $(BIN_DIR)/$* goalert

$(BIN_DIR)/%: go.mod go.sum $(shell find ./devtools -name '*.go')
	go build $(BUILD_FLAGS) -o $@ $(shell find ./devtools -type d -name $* | grep cmd || find ./devtools -type d -name $*)

$(BIN_DIR)/integration/goalert/cypress.json: web/src/cypress.json
	sed 's/\.ts/\.js/' web/src/cypress.json >bin/integration/goalert/cypress.json

$(BIN_DIR)/integration/goalert/cypress: web/src/node_modules web/src/webpack.cypress.js $(BIN_DIR)/integration/goalert/cypress.json $(shell find ./web/src/cypress)
	rm -rf $@
	(cd web/src && yarn webpack --config webpack.cypress.js --target node)
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

install: $(GOFILES)
	go install $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" -ldflags "$(LD_FLAGS)" ./cmd/goalert

cypress: bin/runjson bin/waitfor bin/procwrap bin/simpleproxy bin/mockslack bin/goalert bin/psql-lite web/src/node_modules web/src/schema.d.ts
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

web/src/schema.d.ts: graphql2/schema.graphql web/src/node_modules web/src/genschema.go
	go generate ./web/src

start: bin/waitfor web/src/node_modules web/src/build/vendorPackages.dll.js bin/runjson web/src/schema.d.ts
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	make bin/goalert BUILD_TAGS+=sql_highlight
	GOALERT_VERSION=$(GIT_VERSION) bin/runjson <devtools/runjson/localdev.json

start-prod: bin/waitfor web/inline_data_gen.go bin/runjson
	# force rebuild to ensure build-flags are set
	touch cmd/goalert/main.go
	make bin/goalert BUILD_TAGS+=sql_highlight BUNDLE=1
	bin/runjson <devtools/runjson/localdev-prod.json

jest: web/src/node_modules
	cd web/src && node_modules/.bin/jest $(JEST_ARGS)

test: web/src/node_modules jest
	go test -short ./...

check: generate web/src/node_modules
	# go run devtools/ordermigrations/main.go -check
	go vet ./...
	go run github.com/gordonklaus/ineffassign .
	CGO_ENABLED=0 go run honnef.co/go/tools/cmd/staticcheck ./...
	(cd web/src && yarn run check)
	./devtools/ci/tasks/scripts/codecheck.sh

check-all: check test smoketest cy-wide-prod-run cy-mobile-prod-run

migrate/inline_data_gen.go: migrate/migrations migrate/migrations/*.sql $(INLINER)
	go generate ./migrate

graphql2/mapconfig.go: $(CFGPARAMS) config/config.go graphql2/generated.go devtools/configparams/main.go
	(cd ./graphql2 && go run ../devtools/configparams/main.go -out mapconfig.go && go run golang.org/x/tools/cmd/goimports -w ./mapconfig.go) || go generate ./graphql2

graphql2/maplimit.go: $(CFGPARAMS) limit/id.go graphql2/generated.go devtools/limitapigen/main.go
	(cd ./graphql2 && go run ../devtools/limitapigen/main.go -out maplimit.go && go run golang.org/x/tools/cmd/goimports -w ./maplimit.go) || go generate ./graphql2

graphql2/generated.go: graphql2/schema.graphql graphql2/gqlgen.yml go.mod
	go generate ./graphql2

generate: web/src/node_modules
	go generate ./...

smoketest:
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

web/src/build/static/app.js: web/src/webpack.prod.config.js web/src/yarn.lock $(shell find ./web/src/app -type f ) web/src/schema.d.ts
	rm -rf web/src/build/static
	(cd web/src && yarn --no-progress --silent --frozen-lockfile && node_modules/.bin/webpack --config webpack.prod.config.js --env.GOALERT_VERSION=$(GIT_VERSION))

web/inline_data_gen.go: web/src/build/static/app.js web/src/webpack.prod.config.js $(CFGPARAMS) $(INLINER)
	go generate ./web

web/src/build/vendorPackages.dll.js: web/src/node_modules web/src/webpack.dll.config.js
	(cd web/src && node_modules/.bin/webpack --config ./webpack.dll.config.js --progress)

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
		postgres:11-alpine || docker start goalert-postgres

regendb: bin/resetdb bin/goalert migrate/inline_data_gen.go config.json.bak
	./bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000000
	test -f config.json.bak && bin/goalert set-config --allow-empty-data-encryption-key "--db-url=$(DB_URL)" <config.json.bak || true
	bin/goalert add-user --user-id=00000000-0000-0000-0000-000000000000 --user admin --pass admin123 "--db-url=$(DB_URL)"

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
