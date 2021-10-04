module github.com/target/goalert

go 1.17

require (
	cloud.google.com/go v0.82.0
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/stackdriver v0.13.6
	github.com/99designs/gqlgen v0.13.0
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/agnivade/levenshtein v1.1.0 // indirect
	github.com/alecthomas/chroma v0.9.1
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alexeyco/simpletable v1.0.0
	github.com/aws/aws-sdk-go v1.38.41 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/color v1.11.0
	github.com/felixge/httpsnoop v1.0.2
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/mock v1.5.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.2.0
	github.com/gordonklaus/ineffassign v0.0.0-20210225214923-2e10b2664254
	github.com/gorilla/pat v1.0.1 // indirect
	github.com/graphql-go/graphql v0.7.9
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/yamux v0.0.0-20210707203944-259a57b3608c // pinned version - see https://github.com/target/goalert/issues/1239
	github.com/ian-kent/envconf v0.0.0-20141026121121-c19809918c02 // indirect
	github.com/ian-kent/go-log v0.0.0-20160113211217-5731446c36ab // indirect
	github.com/ian-kent/goose v0.0.0-20141221090059-c3541ea826ad // indirect
	github.com/ian-kent/linkio v0.0.0-20170807205755-97566b872887 // indirect
	github.com/jackc/pgconn v1.8.1
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgtype v1.7.0
	github.com/jackc/pgx/v4 v4.11.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/joho/godotenv v1.3.0
	github.com/mailhog/MailHog v1.0.1
	github.com/mailhog/MailHog-Server v1.0.1
	github.com/mailhog/MailHog-UI v1.0.1 // indirect
	github.com/mailhog/data v1.0.1
	github.com/mailhog/http v1.0.1 // indirect
	github.com/mailhog/mhsendmail v0.2.0 // indirect
	github.com/mailhog/smtp v1.0.1 // indirect
	github.com/mailhog/storage v1.0.1
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/pelletier/go-toml v1.9.1
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.10.0
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20210408115534-a32ed26c37ea
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.9.4
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/t-k/fluent-logger-golang v1.0.0 // indirect
	github.com/tinylib/msgp v1.1.5 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/vbauerster/mpb/v4 v4.12.2
	github.com/vektah/gqlparser/v2 v2.1.0
	go.opencensus.io v0.23.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	golang.org/x/tools v0.1.3
	google.golang.org/genproto v0.0.0-20210624195500-8bfb893ecb84 // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	honnef.co/go/tools v0.1.4
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/puddle v1.1.3 // indirect
	github.com/jaytaylor/html2text v0.0.0-20180606194806-57d518f124b0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/vanng822/css v0.0.0-20190504095207-a21e860bcd04 // indirect
	github.com/vanng822/go-premailer v0.0.0-20191214114701-be27abe028fe // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.46.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
