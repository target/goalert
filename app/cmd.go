package app

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	sloglogrus "github.com/samber/slog-logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/config"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/remotemonitor"
	"github.com/target/goalert/swo"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqldrv"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/version"
	"github.com/target/goalert/web"
	"golang.org/x/term"
)

var shutdownSignalCh = make(chan os.Signal, 2)

// ErrDBRequired is returned when the DB URL is unset.
var ErrDBRequired = validation.NewFieldError("db-url", "is required")

func init() {
	if testing.Testing() {
		// Skip signal handling in tests.
		return
	}

	signal.Notify(shutdownSignalCh, shutdownSignals...)
}

func isCfgNotFound(err error) bool {
	var cfgErr viper.ConfigFileNotFoundError
	return errors.As(err, &cfgErr)
}

// RootCmd is the configuration for running the app binary.
var RootCmd = &cobra.Command{
	Use:   "goalert",
	Short: "Alerting platform.",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := log.FromContext(cmd.Context())

		// update JSON output first
		if viper.GetBool("json") {
			l.EnableJSON()
		}
		if viper.GetBool("verbose") {
			l.EnableDebug()
		}
		if viper.GetBool("log-errors-only") {
			l.ErrorsOnly()
		}

		if viper.GetBool("list-experimental") {
			fmt.Print(`Usage: goalert --experimental=<flag1>,<flag2> ...

These flags are not guaranteed to be stable and may change or be removed at any
time. They are used to enable in-development features and are not intended for 
production use.


Available Flags:

`)

			for _, f := range expflag.AllFlags() {
				fmt.Printf("\t%s\t\t%s", f, expflag.Description(f))
			}

			return nil
		}

		err := viper.ReadInConfig()
		// ignore file not found error
		if err != nil && !isCfgNotFound(err) {
			return errors.Wrap(err, "read config")
		}

		err = initPromServer()
		if err != nil {
			return err
		}
		err = initPprofServer()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		cfg, err := getConfig(ctx)
		if err != nil {
			return err
		}

		doMigrations := func(url string) error {
			if cfg.APIOnly {
				err = migrate.VerifyAll(log.WithDebug(ctx), url)
				if err != nil {
					return errors.Wrap(err, "verify migrations")
				}
				return nil
			}

			s := time.Now()
			n, err := migrate.ApplyAll(log.WithDebug(ctx), url)
			if err != nil {
				return errors.Wrap(err, "apply migrations")
			}
			if n > 0 {
				log.Logf(ctx, "Applied %d migrations in %s.", n, time.Since(s))
			}

			return nil
		}

		err = doMigrations(cfg.DBURL)
		if err != nil {
			return err
		}

		var pool *pgxpool.Pool
		if cfg.DBURLNext != "" {
			err = migrate.VerifyIsLatest(ctx, cfg.DBURL)
			if err != nil {
				return errors.Wrap(err, "verify db")
			}

			err = doMigrations(cfg.DBURLNext)
			if err != nil {
				return errors.Wrap(err, "nextdb")
			}

			mgr, err := swo.NewManager(swo.Config{
				OldDBURL: cfg.DBURL,
				NewDBURL: cfg.DBURLNext,
				CanExec:  !cfg.APIOnly,
				Logger:   cfg.LegacyLogger,
				MaxOpen:  cfg.DBMaxOpen,
				MaxIdle:  cfg.DBMaxIdle,
			})
			if err != nil {
				return errors.Wrap(err, "init switchover handler")
			}
			pool = mgr.Pool()
			cfg.SWO = mgr
		} else {
			appURL, err := sqldrv.AppURL(cfg.DBURL, fmt.Sprintf("GoAlert %s", version.GitVersion()))
			if err != nil {
				return errors.Wrap(err, "connect to postgres")
			}

			poolCfg, err := pgxpool.ParseConfig(appURL)
			if err != nil {
				return errors.Wrap(err, "parse db URL")
			}
			poolCfg.MaxConns = int32(cfg.DBMaxOpen)
			poolCfg.MinConns = int32(cfg.DBMaxIdle)
			sqldrv.SetConfigRetries(poolCfg)

			pool, err = pgxpool.NewWithConfig(context.Background(), poolCfg)
			if err != nil {
				return errors.Wrap(err, "connect to postgres")
			}
		}

		app, err := NewApp(cfg, pool)
		if err != nil {
			return errors.Wrap(err, "init app")
		}

		go handleShutdown(ctx, app.Shutdown)

		// trigger engine cycles by process signal
		trigCh := make(chan os.Signal, 1)
		signal.Notify(trigCh, triggerSignals...)
		go func() {
			for range trigCh {
				app.Trigger()
			}
		}()

		return errors.Wrap(app.Run(ctx), "run app")
	},
}

func handleShutdown(ctx context.Context, fn func(ctx context.Context) error) {
	<-shutdownSignalCh
	log.Logf(ctx, "Application attempting graceful shutdown.")
	sCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	go func() {
		<-shutdownSignalCh
		log.Logf(ctx, "Second signal received, terminating immediately")
		cancel()
	}()

	err := fn(sCtx)
	if err != nil {
		log.Log(ctx, err)
	}
}

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Output the current version.",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations := migrate.Names()

			fmt.Printf(`Version:   %s
GitCommit: %s (%s)
BuildDate: %s
GoVersion: %s (%s)
Platform:  %s/%s
Migration: %s (#%d)
`, version.GitVersion(),
				version.GitCommit(), version.GitTreeState(),
				version.BuildDate().Local().Format(time.RFC3339),
				runtime.Version(), runtime.Compiler,
				runtime.GOOS, runtime.GOARCH,
				migrations[len(migrations)-1], len(migrations),
			)

			return nil
		},
	}

	testCmd = &cobra.Command{
		Use:   "self-test",
		Short: "test suite to validate functionality of GoAlert environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			offlineOnly, _ := cmd.Flags().GetBool("offline")

			var failed bool
			result := func(name string, err error) {
				if err != nil {
					failed = true
					fmt.Printf("%s: FAIL (%v)\n", name, err)
					return
				}
				fmt.Printf("%s: OK\n", name)
			}

			// only do version check if UI is bundled
			if web.AppVersion() != "" {
				var err error
				if version.GitVersion() != web.AppVersion() {
					err = errors.Errorf(
						"mismatch: backend version = '%s'; bundled UI version = '%s'",
						version.GitVersion(),
						web.AppVersion(),
					)
				}
				result("Version", err)
			}

			cf, err := getConfig(cmd.Context())
			if errors.Is(err, ErrDBRequired) {
				err = nil
			}
			if err != nil {
				return err
			}
			var cfg config.Config
			loadConfigDB := func() error {
				conn, err := sql.Open("pgx", cf.DBURL)
				if err != nil {
					return fmt.Errorf("open db: %w", err)
				}

				ctx := cmd.Context()

				storeCfg := config.StoreConfig{
					DB:   conn,
					Keys: cf.EncryptionKeys,
				}
				store, err := config.NewStore(ctx, storeCfg)
				if err != nil {
					return fmt.Errorf("read config: %w", err)
				}
				cfg = store.Config()
				return store.Shutdown(ctx)
			}
			if cf.DBURL != "" && !offlineOnly {
				result("DB", loadConfigDB())
			}

			type service struct {
				name, baseUrl string
			}

			serviceList := []service{
				{name: "Twilio", baseUrl: "https://api.twilio.com/2010-04-01"},
				{name: "Mailgun", baseUrl: "https://api.mailgun.net/v3"},
				{name: "Slack", baseUrl: "https://slack.com/api/api.test"},
			}

			if cfg.OIDC.Enable {
				serviceList = append(serviceList, service{name: "OIDC", baseUrl: cfg.OIDC.IssuerURL + "/.well-known.openid-configuration"})
			}

			if cfg.GitHub.Enable {
				url := "https://github.com"
				if cfg.GitHub.EnterpriseURL != "" {
					url = cfg.GitHub.EnterpriseURL
				}
				serviceList = append(serviceList, service{name: "GitHub", baseUrl: url})
			}

			if offlineOnly {
				serviceList = nil
			}

			for _, s := range serviceList {
				resp, err := http.Get(s.baseUrl)
				result(s.name, err)
				if err == nil {
					resp.Body.Close()
				}
			}

			dstCheck := func() error {
				const (
					standardOffset = -21600
					daylightOffset = -18000
				)
				loc, err := util.LoadLocation("America/Chicago")
				if err != nil {
					return fmt.Errorf("load location: %w", err)
				}
				t := time.Date(2020, time.March, 8, 0, 0, 0, 0, loc)
				_, offset := t.Zone()
				if offset != standardOffset {
					return errors.Errorf("invalid offset: got %d; want %d", offset, standardOffset)
				}
				t = t.Add(3 * time.Hour)
				_, offset = t.Zone()
				if offset != daylightOffset {
					return errors.Errorf("invalid offset: got %d; want %d", offset, daylightOffset)
				}
				t = time.Date(2020, time.November, 1, 0, 0, 0, 0, loc)
				_, offset = t.Zone()
				if offset != daylightOffset {
					return errors.Errorf("invalid offset: got %d; want %d", offset, daylightOffset)
				}
				t = t.Add(3 * time.Hour)
				_, offset = t.Zone()
				if offset != standardOffset {
					return errors.Errorf("invalid offset: got %d; want %d", offset, standardOffset)
				}
				return nil
			}

			result("DST Rules", dstCheck())

			if failed {
				cmd.SilenceUsage = true
				return errors.New("one or more checks failed.")
			}
			return nil
		},
	}

	monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Start a remote-monitoring process that functionally tests alerts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			file := viper.GetString("config-file")
			if file == "" {
				return errors.New("config file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			var cfg remotemonitor.Config
			err = toml.Unmarshal(data, &cfg)
			if err != nil {
				return err
			}

			err = initPromServer()
			if err != nil {
				return err
			}

			err = initPprofServer()
			if err != nil {
				return err
			}

			mon, err := remotemonitor.NewMonitor(cfg)
			if err != nil {
				return err
			}

			handleShutdown(context.Background(), mon.Shutdown)
			return nil
		},
	}

	reEncryptCmd = &cobra.Command{
		Use:   "re-encrypt",
		Short: "Re-encrypt all keyring secrets and config with the current data-encryption-key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.FromContext(cmd.Context())
			// update JSON output first
			if viper.GetBool("json") {
				l.EnableJSON()
			}
			if viper.GetBool("verbose") {
				l.EnableDebug()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if err != nil && !isCfgNotFound(err) {
				return errors.Wrap(err, "read config")
			}

			ctx := cmd.Context()
			c, err := getConfig(ctx)
			if err != nil {
				return err
			}

			if viper.GetString("data-encryption-key") == "" && !viper.GetBool("allow-empty-data-encryption-key") {
				fmt.Println("what", c.EncryptionKeys)
				return validation.NewFieldError("data-encryption-key", "Must not be empty, or set --allow-empty-data-encryption-key")
			}

			db, err := sql.Open("pgx", c.DBURL)
			if err != nil {
				return errors.Wrap(err, "connect to postgres")
			}
			defer db.Close()

			ctx = permission.SystemContext(ctx, "ReEncryptAll")

			return keyring.ReEncryptAll(ctx, db, c.EncryptionKeys)
		},
	}

	exportCmd = &cobra.Command{
		Use:   "export-migrations",
		Short: "Export all migrations as .sql files. Use --export-dir to control the destination.",

		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.FromContext(cmd.Context())
			// update JSON output first
			if viper.GetBool("json") {
				l.EnableJSON()
			}
			if viper.GetBool("verbose") {
				l.EnableDebug()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if err != nil && !isCfgNotFound(err) {
				return errors.Wrap(err, "read config")
			}

			return migrate.DumpMigrations(viper.GetString("export-dir"))
		},
	}

	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform migration(s), then exit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.FromContext(cmd.Context())
			if viper.GetBool("verbose") {
				l.EnableDebug()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if err != nil && !isCfgNotFound(err) {
				return errors.Wrap(err, "read config")
			}

			ctx := cmd.Context()
			c, err := getConfig(ctx)
			if err != nil {
				return err
			}

			down := viper.GetString("down")
			up := viper.GetString("up")
			if down != "" {
				n, err := migrate.Down(ctx, c.DBURL, down)
				if err != nil {
					return errors.Wrap(err, "apply DOWN migrations")
				}
				if n > 0 {
					log.Debugf(ctx, "Applied %d DOWN migrations.", n)
				}
			}

			if up != "" || down == "" {
				n, err := migrate.Up(ctx, c.DBURL, up)
				if err != nil {
					return errors.Wrap(err, "apply UP migrations")
				}
				if n > 0 {
					log.Debugf(ctx, "Applied %d UP migrations.", n)
				}
			}

			return nil
		},
	}

	setConfigCmd = &cobra.Command{
		Use:   "set-config",
		Short: "Sets current config values in the DB from stdin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetString("data-encryption-key") == "" && !viper.GetBool("allow-empty-data-encryption-key") {
				return validation.NewFieldError("data-encryption-key", "Must not be empty, or set --allow-empty-data-encryption-key")
			}
			var data []byte
			if viper.GetString("data") != "" {
				data = []byte(viper.GetString("data"))
			} else {
				if term.IsTerminal(int(os.Stdin.Fd())) {
					// Only print message if we're not piping
					fmt.Println("Enter or paste config data (JSON), then press CTRL+D when done or CTRL+C to quit.")
				}
				intCh := make(chan os.Signal, 1)
				doneCh := make(chan struct{})
				signal.Notify(intCh, os.Interrupt)
				go func() {
					select {
					case <-intCh:
						os.Exit(1)
					case <-doneCh:
					}
				}()

				var err error
				data, err = io.ReadAll(os.Stdin)
				close(doneCh)
				if err != nil {
					return errors.Wrap(err, "read stdin")
				}
			}

			return getSetConfig(cmd.Context(), true, data)
		},
	}

	getConfigCmd = &cobra.Command{
		Use:   "get-config",
		Short: "Gets current config values.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSetConfig(cmd.Context(), false, nil)
		},
	}

	addUserCmd = &cobra.Command{
		Use:   "add-user",
		Short: "Adds a user for basic authentication.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := log.FromContext(cmd.Context())
			if viper.GetBool("verbose") {
				l.EnableDebug()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if err != nil && !isCfgNotFound(err) {
				return errors.Wrap(err, "read config")
			}

			c, err := getConfig(cmd.Context())
			if err != nil {
				return err
			}
			db, err := sql.Open("pgx", c.DBURL)
			if err != nil {
				return errors.Wrap(err, "connect to postgres")
			}
			defer db.Close()

			ctx := permission.SystemContext(cmd.Context(), "AddUser")

			basicStore, err := basic.NewStore(ctx, db)
			if err != nil {
				return errors.Wrap(err, "init basic auth store")
			}

			pass := cmd.Flag("pass").Value.String()
			id := cmd.Flag("user-id").Value.String()
			username := cmd.Flag("user").Value.String()

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "begin tx")
			}
			defer sqlutil.Rollback(ctx, "add-user", tx)

			if id == "" {
				u := &user.User{
					Name:  username,
					Email: cmd.Flag("email").Value.String(),
					Role:  permission.RoleUser,
				}
				if cmd.Flag("admin").Value.String() == "true" {
					u.Role = permission.RoleAdmin
				}
				userStore, err := user.NewStore(ctx, db)
				if err != nil {
					return errors.Wrap(err, "init user store")
				}
				u, err = userStore.InsertTx(ctx, tx, u)
				if err != nil {
					return errors.Wrap(err, "create user")
				}
				id = u.ID
			}

			if pass == "" {
				fmt.Fprint(os.Stderr, "New Password: ")
				p, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return errors.Wrap(err, "get password")
				}
				pass = string(p)
				fmt.Fprintln(os.Stderr)
			}

			pw, err := basicStore.NewHashedPassword(ctx, pass)
			if err != nil {
				return errors.Wrap(err, "hash password")
			}

			err = basicStore.CreateTx(ctx, tx, id, username, pw)
			if err != nil {
				return errors.Wrap(err, "add basic auth entry")
			}

			err = tx.Commit()
			if err != nil {
				return errors.Wrap(err, "commit tx")
			}

			log.Logf(ctx, "Username '%s' added.", username)

			return nil
		},
	}
)

// getConfig will load the current configuration from viper
func getConfig(ctx context.Context) (Config, error) {
	cfg := Config{
		LegacyLogger: log.FromContext(ctx),

		JSON:        viper.GetBool("json"),
		LogRequests: viper.GetBool("log-requests"),
		LogEngine:   viper.GetBool("log-engine-cycles"),
		Verbose:     viper.GetBool("verbose"),
		APIOnly:     viper.GetBool("api-only"),

		DBMaxOpen: viper.GetInt("db-max-open"),
		DBMaxIdle: viper.GetInt("db-max-idle"),

		PublicURL: viper.GetString("public-url"),

		MaxReqBodyBytes:   viper.GetInt64("max-request-body-bytes"),
		MaxReqHeaderBytes: viper.GetInt("max-request-header-bytes"),

		DisableHTTPSRedirect: viper.GetBool("disable-https-redirect"),
		EnableSecureHeaders:  viper.GetBool("enable-secure-headers"),

		ListenAddr: viper.GetString("listen"),

		TLSListenAddr: viper.GetString("listen-tls"),

		SysAPIListenAddr: viper.GetString("listen-sysapi"),
		SysAPICertFile:   viper.GetString("sysapi-cert-file"),
		SysAPIKeyFile:    viper.GetString("sysapi-key-file"),
		SysAPICAFile:     viper.GetString("sysapi-ca-file"),

		SMTPListenAddr:        viper.GetString("smtp-listen"),
		SMTPListenAddrTLS:     viper.GetString("smtp-listen-tls"),
		SMTPAdditionalDomains: viper.GetString("smtp-additional-domains"),
		SMTPMaxRecipients:     viper.GetInt("smtp-max-recipients"),

		EmailIntegrationDomain: viper.GetString("email-integration-domain"),

		EngineCycleTime: viper.GetDuration("engine-cycle-time"),

		HTTPPrefix: viper.GetString("http-prefix"),

		SlackBaseURL:  viper.GetString("slack-base-url"),
		TwilioBaseURL: viper.GetString("twilio-base-url"),

		DBURL:     viper.GetString("db-url"),
		DBURLNext: viper.GetString("db-url-next"),

		StatusAddr: viper.GetString("status-addr"),

		EncryptionKeys: keyring.Keys{[]byte(viper.GetString("data-encryption-key")), []byte(viper.GetString("data-encryption-key-old"))},

		RegionName: viper.GetString("region-name"),

		StubNotifiers: viper.GetBool("stub-notifiers"),

		UIDir: viper.GetString("ui-dir"),
	}

	opts := sloglogrus.Option{
		Level:  slog.LevelInfo,
		Logger: cfg.LegacyLogger.Logrus(),
	}
	if viper.GetBool("log-errors-only") {
		opts.Level = slog.LevelError
	} else if cfg.Verbose {
		opts.Level = slog.LevelDebug
	}
	cfg.Logger = slog.New(opts.NewLogrusHandler())

	var fs expflag.FlagSet
	strict := viper.GetBool("strict-experimental")
	s := viper.GetStringSlice("experimental")
	if len(s) == 1 {
		s = strings.Split(s[0], ",")
	}
	for _, f := range s {
		if strict && expflag.Description(expflag.Flag(f)) == "" {
			return cfg, errors.Errorf("unknown experimental flag: %s", f)
		}

		fs = append(fs, expflag.Flag(f))
	}
	cfg.ExpFlags = fs

	if cfg.PublicURL != "" {
		u, err := url.Parse(cfg.PublicURL)
		if err != nil {
			return cfg, errors.Wrap(err, "parse public url")
		}
		if u.Scheme == "" {
			return cfg, errors.New("public-url must be an absolute URL (missing scheme)")
		}
		u.Path = strings.TrimSuffix(u.Path, "/")
		cfg.PublicURL = u.String()
		if cfg.HTTPPrefix != "" {
			return cfg, errors.New("public-url and http-prefix cannot be used together")
		}
		cfg.HTTPPrefix = u.Path
	}

	if cfg.DBURL == "" {
		return cfg, ErrDBRequired
	}

	var err error
	cfg.TLSConfig, err = getTLSConfig("")
	if err != nil {
		return cfg, err
	}
	if cfg.TLSConfig != nil {
		cfg.TLSConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	if cfg.SMTPListenAddr != "" || cfg.SMTPListenAddrTLS != "" {
		if cfg.EmailIntegrationDomain == "" {
			return cfg, errors.New("email-integration-domain is required when smtp-listen or smtp-listen-tls is set")
		}
	}

	cfg.TLSConfigSMTP, err = getTLSConfig("smtp-")
	if err != nil {
		return cfg, err
	}

	if viper.GetBool("stack-traces") {
		log.FromContext(ctx).EnableStacks()
	}
	return cfg, nil
}

func init() {
	def := Defaults()
	RootCmd.Flags().StringP("listen", "l", def.ListenAddr, "Listen address:port for the application.")

	RootCmd.Flags().StringP("listen-tls", "t", def.TLSListenAddr, "HTTPS listen address:port for the application.  Requires setting --tls-cert-data and --tls-key-data OR --tls-cert-file and --tls-key-file.")

	RootCmd.Flags().StringSlice("experimental", nil, "Enable experimental features.")
	RootCmd.Flags().Bool("list-experimental", false, "List experimental features.")
	RootCmd.Flags().Bool("strict-experimental", false, "Fail to start if unknown experimental features are specified.")

	RootCmd.Flags().String("listen-sysapi", "", "(Experimental) Listen address:port for the system API (gRPC).")

	RootCmd.Flags().String("sysapi-cert-file", "", "(Experimental) Specifies a path to a PEM-encoded certificate to use when connecting to plugin services.")
	RootCmd.Flags().String("sysapi-key-file", "", "(Experimental) Specifies a path to a PEM-encoded private key file use when connecting to plugin services.")
	RootCmd.Flags().String("sysapi-ca-file", "", "(Experimental) Specifies a path to a PEM-encoded certificate(s) to authorize connections from plugin services.")

	RootCmd.Flags().String("public-url", "", "Externally routable URL to the application. Used for validating callback requests, links, auth, and prefix calculation.")

	RootCmd.PersistentFlags().StringP("listen-prometheus", "p", "", "Bind address for Prometheus metrics.")
	RootCmd.PersistentFlags().String("listen-pprof", "", "Bind address for pprof.")
	RootCmd.PersistentFlags().Int("pprof-block-profile-rate", 0, "Set the block profile rate in hz.")
	RootCmd.PersistentFlags().Int("pprof-mutex-profile-fraction", 0, "Set the mutex profile fraction (rate is 1/this-value).")

	RootCmd.Flags().String("tls-cert-file", "", "Specifies a path to a PEM-encoded certificate.  Has no effect if --listen-tls is unset.")
	RootCmd.Flags().String("tls-key-file", "", "Specifies a path to a PEM-encoded private key file.  Has no effect if --listen-tls is unset.")
	RootCmd.Flags().String("tls-cert-data", "", "Specifies a PEM-encoded certificate.  Has no effect if --listen-tls is unset.")
	RootCmd.Flags().String("tls-key-data", "", "Specifies a PEM-encoded private key.  Has no effect if --listen-tls is unset.")

	RootCmd.Flags().String("smtp-listen", "", "Listen address:port for an internal SMTP server.")
	RootCmd.Flags().String("smtp-listen-tls", "", "SMTPS listen address:port for an internal SMTP server.  Requires setting --smtp-tls-cert-data and --smtp-tls-key-data OR --smtp-tls-cert-file and --smtp-tls-key-file.")
	RootCmd.Flags().String("email-integration-domain", "", "This flag is required to set the domain used for email integration keys when --smtp-listen or --smtp-listen-tls are set.")

	RootCmd.Flags().String("smtp-tls-cert-file", "", "Specifies a path to a PEM-encoded certificate.  Has no effect if --smtp-listen-tls is unset.")
	RootCmd.Flags().String("smtp-tls-key-file", "", "Specifies a path to a PEM-encoded private key file.  Has no effect if --smtp-listen-tls is unset.")
	RootCmd.Flags().String("smtp-tls-cert-data", "", "Specifies a PEM-encoded certificate.  Has no effect if --smtp-listen-tls is unset.")
	RootCmd.Flags().String("smtp-tls-key-data", "", "Specifies a PEM-encoded private key.  Has no effect if --smtp-listen-tls is unset.")

	RootCmd.Flags().Int("smtp-max-recipients", def.SMTPMaxRecipients, "Specifies the maximum number of recipients allowed per message.")
	RootCmd.Flags().String("smtp-additional-domains", "", "Specifies additional destination domains that are allowed for the SMTP server.  For multiple domains, separate them with a comma, e.g., \"domain1.com,domain2.org,domain3.net\".")

	RootCmd.Flags().Duration("engine-cycle-time", def.EngineCycleTime, "Time between engine cycles.")

	RootCmd.Flags().String("http-prefix", def.HTTPPrefix, "Specify the HTTP prefix of the application.")
	_ = RootCmd.Flags().MarkDeprecated("http-prefix", "use --public-url instead")

	RootCmd.Flags().Bool("api-only", def.APIOnly, "Starts in API-only mode (schedules & notifications will not be processed). Useful in clusters.")

	RootCmd.Flags().Int("db-max-open", def.DBMaxOpen, "Max open DB connections.")
	RootCmd.Flags().Int("db-max-idle", def.DBMaxIdle, "Max idle DB connections.")

	RootCmd.Flags().Int64("max-request-body-bytes", def.MaxReqBodyBytes, "Max body size for all incoming requests (in bytes). Set to 0 to disable limit.")
	RootCmd.Flags().Int("max-request-header-bytes", def.MaxReqHeaderBytes, "Max header size for all incoming requests (in bytes). Set to 0 to disable limit.")

	// No longer used
	RootCmd.Flags().String("github-base-url", "", "Base URL for GitHub auth and API calls.")

	RootCmd.Flags().String("twilio-base-url", def.TwilioBaseURL, "Override the Twilio API URL.")
	RootCmd.Flags().String("slack-base-url", def.SlackBaseURL, "Override the Slack base URL.")

	RootCmd.Flags().String("region-name", def.RegionName, "Name of region for message processing (case sensitive). Only one instance per-region-name will process outgoing messages.")

	RootCmd.PersistentFlags().String("db-url", def.DBURL, "Connection string for Postgres.")
	RootCmd.PersistentFlags().String("db-url-next", def.DBURLNext, "Connection string for the *next* Postgres server (enables DB switchover mode).")

	RootCmd.Flags().String("jaeger-endpoint", "", "Jaeger HTTP Thrift endpoint")
	RootCmd.Flags().String("jaeger-agent-endpoint", "", "Instructs Jaeger exporter to send spans to jaeger-agent at this address.")
	_ = RootCmd.Flags().MarkDeprecated("jaeger-endpoint", "Jaeger support has been removed.")
	_ = RootCmd.Flags().MarkDeprecated("jaeger-agent-endpoint", "Jaeger support has been removed.")
	RootCmd.Flags().String("stackdriver-project-id", "", "Project ID for Stackdriver. Enables tracing output to Stackdriver.")
	_ = RootCmd.Flags().MarkDeprecated("stackdriver-project-id", "Stackdriver support has been removed.")
	RootCmd.Flags().String("tracing-cluster-name", "", "Cluster name to use for tracing (i.e. kubernetes, Stackdriver/GKE environment).")
	_ = RootCmd.Flags().MarkDeprecated("tracing-cluster-name", "Tracing support has been removed.")
	RootCmd.Flags().String("tracing-pod-namespace", "", "Pod namespace to use for tracing.")
	_ = RootCmd.Flags().MarkDeprecated("tracing-pod-namespace", "Tracing support has been removed.")
	RootCmd.Flags().String("tracing-pod-name", "", "Pod name to use for tracing.")
	_ = RootCmd.Flags().MarkDeprecated("tracing-pod-name", "Tracing support has been removed.")
	RootCmd.Flags().String("tracing-container-name", "", "Container name to use for tracing.")
	_ = RootCmd.Flags().MarkDeprecated("tracing-container-name", "Tracing support has been removed.")
	RootCmd.Flags().String("tracing-node-name", "", "Node name to use for tracing.")
	_ = RootCmd.Flags().MarkDeprecated("tracing-node-name", "Tracing support has been removed.")
	RootCmd.Flags().Float64("tracing-probability", 0, "Probability of a new trace to be recorded.")
	_ = RootCmd.Flags().MarkDeprecated("tracing-probability", "Tracing support has been removed.")

	RootCmd.Flags().Duration("kubernetes-cooldown", 0, "Cooldown period, from the last TCP connection, before terminating the listener when receiving a shutdown signal.")
	_ = RootCmd.Flags().MarkDeprecated("kubernetes-cooldown", "Use lifecycle hooks (preStop) instead.")
	RootCmd.Flags().String("status-addr", def.StatusAddr, "Open a port to emit status updates. Connections are closed when the server shuts down. Can be used to keep containers running until GoAlert has exited.")

	RootCmd.PersistentFlags().String("data-encryption-key", "", "Used to generate an encryption key for sensitive data like signing keys. Can be any length. Only use this when performing a switchover.")
	RootCmd.PersistentFlags().String("data-encryption-key-old", "", "Fallback key. Used for decrypting existing data only. Only necessary when changing --data-encryption-key.")
	RootCmd.PersistentFlags().Bool("stack-traces", false, "Enables stack traces with all error logs.")

	RootCmd.Flags().Bool("stub-notifiers", def.StubNotifiers, "If true, notification senders will be replaced with a stub notifier that always succeeds (useful for staging/sandbox environments).")

	RootCmd.PersistentFlags().BoolP("verbose", "v", def.Verbose, "Enable verbose logging.")
	RootCmd.Flags().Bool("log-requests", def.LogRequests, "Log all HTTP requests. If false, requests will be logged for debug/trace contexts only.")
	RootCmd.Flags().Bool("log-engine-cycles", def.LogEngine, "Log start and end of each engine cycle.")
	RootCmd.PersistentFlags().Bool("json", def.JSON, "Log in JSON format.")
	RootCmd.PersistentFlags().Bool("log-errors-only", false, "Only log errors (superseeds other flags).")

	RootCmd.Flags().String("ui-dir", "", "Serve UI assets from a local directory instead of from memory.")

	RootCmd.Flags().Bool("disable-https-redirect", def.DisableHTTPSRedirect, "Disable automatic HTTPS redirects.")
	RootCmd.Flags().Bool("enable-secure-headers", false, "Enable secure headers (X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, Content-Security-Policy).")

	migrateCmd.Flags().String("up", "", "Target UP migration to apply.")
	migrateCmd.Flags().String("down", "", "Target DOWN migration to roll back to.")
	exportCmd.Flags().String("export-dir", "migrations", "Destination dir for export. If it does not exist, it will be created.")

	addUserCmd.Flags().String("user-id", "", "If specified, the auth entry will be created for an existing user ID. Default is to create a new user.")
	addUserCmd.Flags().String("pass", "", "Specify new users password (if blank, prompt will be given).")
	addUserCmd.Flags().String("user", "", "Specifies the login username.")
	addUserCmd.Flags().String("email", "", "Specifies the email address of the new user (ignored if user-id is provided).")
	addUserCmd.Flags().Bool("admin", false, "If specified, the user will be created with the admin role (ignored if user-id is provided).")

	setConfigCmd.Flags().String("data", "", "Use data instead of reading config from stdin.")
	setConfigCmd.Flags().Bool("allow-empty-data-encryption-key", false, "Explicitly allow an empty data-encryption-key when setting config or re-encrypting data.")
	reEncryptCmd.Flags().AddFlag(setConfigCmd.Flag("allow-empty-data-encryption-key"))

	testCmd.Flags().Bool("offline", false, "Only perform offline checks.")

	monitorCmd.Flags().StringP("config-file", "f", "", "Configuration file for monitoring (required).")
	initCertCommands()
	RootCmd.AddCommand(versionCmd, testCmd, migrateCmd, exportCmd, monitorCmd, addUserCmd, getConfigCmd, setConfigCmd, genCerts, reEncryptCmd)

	err := viper.BindPFlags(RootCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(reEncryptCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(monitorCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(migrateCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(exportCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(setConfigCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(getConfigCmd.Flags())
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}

	viper.SetEnvPrefix("GOALERT")

	// use underscores in env names
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.AutomaticEnv()
}
