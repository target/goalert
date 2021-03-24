module github.com/target/goalert

go 1.16

require (
	cloud.google.com/go v0.74.0
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	dmitri.shuralyov.com/go/generated v0.0.0-20170818220700-b1254a446363 // indirect
	github.com/99designs/gqlgen v0.13.0
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/agnivade/levenshtein v1.1.0 // indirect
	github.com/alecthomas/chroma v0.8.2
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alexeyco/simpletable v0.0.0-20200730140406-5bb24159ccfb
	github.com/aws/aws-sdk-go v1.28.6 // indirect
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/fatih/color v1.10.0
	github.com/felixge/httpsnoop v1.0.1
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/mock v1.4.4
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20201107091007-3b93a8888063
	github.com/gorilla/pat v1.0.1 // indirect
	github.com/graphql-go/graphql v0.7.9
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // pinned version - see https://github.com/target/goalert/issues/1239
	github.com/ian-kent/envconf v0.0.0-20141026121121-c19809918c02 // indirect
	github.com/ian-kent/go-log v0.0.0-20160113211217-5731446c36ab // indirect
	github.com/ian-kent/goose v0.0.0-20141221090059-c3541ea826ad // indirect
	github.com/ian-kent/linkio v0.0.0-20170807205755-97566b872887 // indirect
	github.com/jackc/pgconn v1.7.2
	github.com/jackc/pgtype v1.6.1
	github.com/jackc/pgx/v4 v4.9.2
	github.com/jmespath/go-jmespath v0.4.0
	github.com/joho/godotenv v1.3.0
	github.com/mailhog/MailHog v1.0.1 // indirect
	github.com/mailhog/MailHog-Server v1.0.1
	github.com/mailhog/MailHog-UI v1.0.1 // indirect
	github.com/mailhog/data v1.0.1
	github.com/mailhog/http v1.0.1 // indirect
	github.com/mailhog/mhsendmail v0.2.0 // indirect
	github.com/mailhog/smtp v1.0.1 // indirect
	github.com/mailhog/storage v1.0.1
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mhale/smtpd v0.0.0-20200509114310-d7a07f752336
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/pelletier/go-toml v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/t-k/fluent-logger-golang v1.0.0 // indirect
	github.com/tinylib/msgp v1.1.5 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.1.0
	github.com/ugorji/go v1.1.4 // indirect
	github.com/urfave/cli/v2 v2.2.0 // indirect
	github.com/vbauerster/mpb/v4 v4.12.2
	github.com/vektah/gqlparser/v2 v2.1.0
	go.opencensus.io v0.22.5
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20201216054612-986b41b23924
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	golang.org/x/sys v0.0.0-20201223074533-0d417f636930
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	golang.org/x/tools v0.0.0-20201223174954-9cbb1efa7745
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/square/go-jose.v2 v2.4.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
