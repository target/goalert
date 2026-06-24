module github.com/target/goalert

go 1.26.0

replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.16

require (
	github.com/99designs/gqlgen v0.17.90
	github.com/brianvoe/gofakeit/v7 v7.15.0
	github.com/coreos/go-oidc/v3 v3.18.0
	github.com/creack/pty/v2 v2.0.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/deckarep/golang-set/v2 v2.9.0
	github.com/emersion/go-smtp v0.24.0
	github.com/expr-lang/expr v1.17.8
	github.com/fatih/color v1.19.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8
	github.com/google/go-github/v86 v86.0.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/yamux v0.1.2
	github.com/jackc/pgtype v1.14.4
	github.com/jackc/pgx/v5 v5.9.2
	github.com/jackc/pgxlisten v0.0.0-20250802141604-12b92425684c
	github.com/jmespath/go-jmespath v0.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.12.3
	github.com/matcornic/hermes v1.3.0
	github.com/mnako/letters v0.2.8
	github.com/nyaruka/phonenumbers v1.7.5
	github.com/oauth2-proxy/mockoidc v0.0.0-20240214162133-caebfff84d25
	github.com/pelletier/go-toml/v2 v2.3.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/riverqueue/river v0.38.0
	github.com/riverqueue/river/riverdriver/riverdatabasesql v0.38.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.38.0
	github.com/riverqueue/river/rivertype v0.38.0
	github.com/samber/slog-logrus/v2 v2.5.4
	github.com/sirupsen/logrus v1.9.4
	github.com/slack-go/slack v0.24.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/sqlc-dev/pqtype v0.3.0
	github.com/stretchr/testify v1.11.1
	github.com/vektah/gqlparser/v2 v2.5.33
	golang.org/x/crypto v0.52.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sys v0.45.0
	golang.org/x/term v0.43.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	riverqueue.com/riverui v0.16.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/PuerkitoBio/goquery v1.12.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/cubicdaiya/gonp v1.0.4 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emersion/go-sasl v0.0.0-20241020182733-b788ff22d5a6 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.37.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fullstorydev/grpcui v1.4.3 // indirect
	github.com/fullstorydev/grpcurl v1.9.3 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-jose/go-jose/v3 v3.0.5 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.2 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.28.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v1.0.2 // indirect
	github.com/inbucket/html2text v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20250907135507-afb5586c32a6 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jaytaylor/html2text v0.0.0-20260303211410-1a4bdc82ecec // indirect
	github.com/jhump/protoreflect v1.18.0 // indirect
	github.com/jhump/protoreflect/v2 v2.0.0-beta.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/kffl/speedbump v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.23 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-sqlite3 v0.34.0 // indirect
	github.com/ncruces/go-sqlite3-wasm/v2 v2.2.35300 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/errors v1.3.0 // indirect
	github.com/olekukonko/ll v0.1.8 // indirect
	github.com/olekukonko/tablewriter v1.1.4 // indirect
	github.com/petermattis/goid v0.0.0-20260330135022-df67b199bc81 // indirect
	github.com/pganalyze/pg_query_go/v6 v6.2.2 // indirect
	github.com/pingcap/errors v0.11.5-0.20250523034308-74f78ae071ee // indirect
	github.com/pingcap/failpoint v0.0.0-20260406204437-bbc9d102c19e // indirect
	github.com/pingcap/log v1.1.0 // indirect
	github.com/pingcap/tidb/pkg/parser v0.0.0-20260504140133-511dba1dbe17 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/riverqueue/apiframe v0.0.0-20260428012848-22cd8d31a740 // indirect
	github.com/riverqueue/river/cmd/river v0.35.1 // indirect
	github.com/riverqueue/river/riverdriver v0.38.0 // indirect
	github.com/riverqueue/river/riverdriver/riversqlite v0.35.1 // indirect
	github.com/riverqueue/river/rivershared v0.38.0 // indirect
	github.com/riza-io/grpc-go v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/samber/slog-common v0.22.0 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/sqlc-dev/doubleclick v1.0.0 // indirect
	github.com/sqlc-dev/sqlc v1.31.1 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/urfave/cli/v3 v3.8.0 // indirect
	github.com/vanng822/css v1.0.1 // indirect
	github.com/vanng822/go-premailer v1.33.0 // indirect
	github.com/wasilibs/go-pgquery v0.0.0-20260428021157-dca720e45577 // indirect
	github.com/wasilibs/wazero-helpers v0.0.0-20250123031827-cd30c44769bb // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/telemetry v0.0.0-20260428171046-76f71b9afea0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260504160031-60b97b32f348 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260504160031-60b97b32f348 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.6.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.72.2 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.50.0 // indirect
)

tool (
	github.com/99designs/gqlgen
	github.com/fullstorydev/grpcui/cmd/grpcui
	github.com/kffl/speedbump
	github.com/riverqueue/river/cmd/river
	github.com/sqlc-dev/sqlc/cmd/sqlc
	github.com/target/goalert/devtools/configparams
	github.com/target/goalert/devtools/configtsidgen
	github.com/target/goalert/devtools/genmake
	github.com/target/goalert/devtools/gettool
	github.com/target/goalert/devtools/gqltsgen
	github.com/target/goalert/devtools/limitapigen
	github.com/target/goalert/devtools/mockoidc
	github.com/target/goalert/devtools/mockslack/cmd/mockslack
	github.com/target/goalert/devtools/ordermigrations
	github.com/target/goalert/devtools/pgdump-lite/cmd/pgdump-lite
	github.com/target/goalert/devtools/pgmocktime/cmd/pgmocktime
	github.com/target/goalert/devtools/procwrap
	github.com/target/goalert/devtools/psql-lite
	github.com/target/goalert/devtools/resetdb
	github.com/target/goalert/devtools/runproc
	github.com/target/goalert/devtools/scripts/db-url-set-db
	github.com/target/goalert/devtools/simpleproxy
	github.com/target/goalert/devtools/waitfor
	github.com/target/goalert/expflag/cmd/expflagtsgen
	github.com/target/goalert/migrate/cmd/goalert-migrate
	golang.org/x/tools/cmd/goimports
	golang.org/x/tools/cmd/stringer
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
	google.golang.org/protobuf/cmd/protoc-gen-go
)
