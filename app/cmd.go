package app

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/stdlib"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/remotemonitor"
	"github.com/target/goalert/sqltrace"
	"github.com/target/goalert/switchover"
	"github.com/target/goalert/switchover/dbsync"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/version"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
)

var shutdownSignalCh = make(chan os.Signal, 2)

func init() {
	signal.Notify(shutdownSignalCh, shutdownSignals...)
}

// RootCmd is the configuration for running the app binary.
var RootCmd = &cobra.Command{
	Use:   "goalert",
	Short: "Alerting platform.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// update JSON output first
		if viper.GetBool("json") {
			log.EnableJSON()
		}
		if viper.GetBool("verbose") {
			log.EnableVerbose()
		}
		if viper.GetBool("log-errors-only") {
			log.ErrorsOnly()
		}

		err := viper.ReadInConfig()
		// ignore file not found error
		if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
			return errors.Wrap(err, "read config")
		}

		ctx := context.Background()
		cfg, err := getConfig()
		if err != nil {
			return err
		}
		exporters, err := configTracing(ctx, cfg)
		if err != nil {
			return errors.Wrap(err, "config tracing")
		}

		defer func() {
			// flush exporters
			type flusher interface {
				Flush()
			}
			for _, e := range exporters {
				if f, ok := e.(flusher); ok {
					f.Flush()
				}
			}
		}()

		wrappedDriver := sqltrace.WrapDriver(stdlib.GetDefaultDriver(), &sqltrace.WrapOptions{Query: true})

		u, err := url.Parse(cfg.DBURL)
		if err != nil {
			return errors.Wrap(err, "parse old URL")
		}
		q := u.Query()
		if cfg.DBURLNext != "" {
			q.Set("application_name", fmt.Sprintf("GoAlert %s (S/O Mode)", version.GitVersion()))
		} else {
			q.Set("application_name", fmt.Sprintf("GoAlert %s", version.GitVersion()))
		}
		u.RawQuery = q.Encode()
		cfg.DBURL = u.String()

		s := time.Now()
		n, err := migrate.ApplyAll(log.EnableDebug(ctx), cfg.DBURL)
		if err != nil {
			return errors.Wrap(err, "apply migrations")
		}
		if n > 0 {
			log.Logf(ctx, "Applied %d migrations in %s.", n, time.Since(s))
		}

		dbc, err := wrappedDriver.OpenConnector(cfg.DBURL)
		if err != nil {
			return errors.Wrap(err, "connect to postgres")
		}
		var db *sql.DB
		var h *switchover.Handler
		if cfg.DBURLNext != "" {
			u, err := url.Parse(cfg.DBURLNext)
			if err != nil {
				return errors.Wrap(err, "parse next URL")
			}
			q := u.Query()
			q.Set("application_name", fmt.Sprintf("GoAlert %s (S/O Mode)", version.GitVersion()))
			u.RawQuery = q.Encode()
			cfg.DBURLNext = u.String()

			dbcNext, err := wrappedDriver.OpenConnector(cfg.DBURLNext)
			if err != nil {
				return errors.Wrap(err, "connect to postres (next)")
			}
			h, err = switchover.NewHandler(ctx, dbc, dbcNext, cfg.DBURL, cfg.DBURLNext)
			if err != nil {
				return errors.Wrap(err, "init changeover handler")
			}
			db = h.DB()
		} else {
			db = sql.OpenDB(dbc)
		}

		app, err := NewApp(cfg, db)
		if err != nil {
			return errors.Wrap(err, "init app")
		}
		if h != nil {
			h.SetApp(app)
		}

		go handleShutdown(ctx, func(ctx context.Context) error {
			if h != nil {
				h.Abort()
			}
			return app.Shutdown(ctx)
		})

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
	sCtx, sp := trace.StartSpan(sCtx, "Shutdown")
	defer sp.End()
	go func() {
		<-shutdownSignalCh
		log.Logf(ctx, "Second signal received, terminating immediately")
		sp.Annotate([]trace.Attribute{trace.BoolAttribute("shutdown.force", true)}, "Second signal received.")
		cancel()
	}()

	err := fn(sCtx)
	if err != nil {
		sp.Annotate([]trace.Attribute{trace.BoolAttribute("error", true)}, err.Error())
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

	switchCmd = &cobra.Command{
		Use:   "switchover-shell",
		Short: "Start a the switchover shell, used to initiate, control, and monitor a DB switchover operation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := getConfig()
			if err != nil {
				return err
			}

			if cfg.DBURLNext == "" {
				return validation.NewFieldError("DBURLNext", "must not be empty for switchover")
			}

			return dbsync.RunShell(cfg.DBURL, cfg.DBURLNext)
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

			t, err := toml.LoadFile(file)
			if err != nil {
				return err
			}

			var cfg remotemonitor.Config
			err = t.Unmarshal(&cfg)
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

	exportCmd = &cobra.Command{
		Use:   "export-migrations",
		Short: "Export all migrations as .sql files. Use --export-dir to control the destination.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// update JSON output first
			if viper.GetBool("json") {
				log.EnableJSON()
			}
			if viper.GetBool("verbose") {
				log.EnableVerbose()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
				return errors.Wrap(err, "read config")
			}

			return migrate.DumpMigrations(viper.GetString("export-dir"))
		},
	}

	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform migration(s), then exit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("verbose") {
				log.EnableVerbose()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
				return errors.Wrap(err, "read config")
			}

			c, err := getConfig()
			if err != nil {
				return err
			}

			ctx := context.Background()
			down := viper.GetString("down")
			up := viper.GetString("up")
			if down != "" {
				n, err := migrate.Down(ctx, c.DBURL, down)

				if err != nil {
					return errors.Wrap(err, "apply DOWN migrations")
				}
				if n > 0 {
					log.Debugf(context.TODO(), "Applied %d DOWN migrations.", n)
				}
			}

			if up != "" || down == "" {
				n, err := migrate.Up(ctx, c.DBURL, up)

				if err != nil {
					return errors.Wrap(err, "apply UP migrations")
				}
				if n > 0 {
					log.Debugf(context.TODO(), "Applied %d UP migrations.", n)
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
				if terminal.IsTerminal(int(os.Stdin.Fd())) {
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
				data, err = ioutil.ReadAll(os.Stdin)
				close(doneCh)
				if err != nil {
					return errors.Wrap(err, "read stdin")
				}
			}

			return getSetConfig(true, data)
		},
	}

	getConfigCmd = &cobra.Command{
		Use:   "get-config",
		Short: "Gets current config values.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSetConfig(false, nil)
		},
	}

	addUserCmd = &cobra.Command{
		Use:   "add-user",
		Short: "Adds a user for basic authentication.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("verbose") {
				log.EnableVerbose()
			}

			err := viper.ReadInConfig()
			// ignore file not found error
			if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
				return errors.Wrap(err, "read config")
			}

			c, err := getConfig()
			if err != nil {
				return err
			}
			db, err := sql.Open("postgres", c.DBURL)
			if err != nil {
				return errors.Wrap(err, "connect to postgres")
			}
			defer db.Close()

			ctx := permission.SystemContext(context.Background(), "AddUser")

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
			defer tx.Rollback()

			if id == "" {
				u := &user.User{
					Name:  username,
					Email: cmd.Flag("email").Value.String(),
					Role:  permission.RoleUser,
				}
				if cmd.Flag("admin").Value.String() == "true" {
					u.Role = permission.RoleAdmin
				}
				userStore, err := user.NewDB(ctx, db)
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
				fmt.Printf("New Password: ")
				p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return errors.Wrap(err, "get password")
				}
				pass = string(p)
				fmt.Printf("\n'%s'\n", pass)
			}

			err = basicStore.CreateTx(ctx, tx, id, username, pass)
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
func getConfig() (appConfig, error) {
	cfg := appConfig{
		JSON:        viper.GetBool("json"),
		LogRequests: viper.GetBool("log-requests"),
		Verbose:     viper.GetBool("verbose"),
		APIOnly:     viper.GetBool("api-only"),

		DBMaxOpen: viper.GetInt("db-max-open"),
		DBMaxIdle: viper.GetInt("db-max-idle"),

		MaxReqBodyBytes:   viper.GetInt64("max-request-body-bytes"),
		MaxReqHeaderBytes: viper.GetInt("max-request-header-bytes"),

		DisableHTTPSRedirect: viper.GetBool("disable-https-redirect"),

		ListenAddr: viper.GetString("listen"),

		SlackBaseURL:  viper.GetString("slack-base-url"),
		TwilioBaseURL: viper.GetString("twilio-base-url"),

		DBURL:     viper.GetString("db-url"),
		DBURLNext: viper.GetString("db-url-next"),

		JaegerEndpoint:      viper.GetString("jaeger-endpoint"),
		JaegerAgentEndpoint: viper.GetString("jaeger-agent-endpoint"),

		StackdriverProjectID: viper.GetString("stackdriver-project-id"),

		TracingClusterName:   viper.GetString("tracing-cluster-name"),
		TracingPodNamespace:  viper.GetString("tracing-pod-namespace"),
		TracingPodName:       viper.GetString("tracing-pod-name"),
		TracingContainerName: viper.GetString("tracing-container-name"),
		TracingNodeName:      viper.GetString("tracing-node-name"),
		TraceProbability:     viper.GetFloat64("tracing-probability"),

		KubernetesCooldown: viper.GetDuration("kubernetes-cooldown"),
		StatusAddr:         viper.GetString("status-addr"),

		EncryptionKeys: keyring.Keys{[]byte(viper.GetString("data-encryption-key")), []byte(viper.GetString("data-encryption-key-old"))},

		RegionName: viper.GetString("region-name"),

		StubNotifiers: viper.GetBool("stub-notifiers"),

		UIURL: viper.GetString("ui-url"),
	}

	if cfg.DBURL == "" {
		return cfg, validation.NewFieldError("db-url", "is required")
	}

	if viper.GetBool("stack-traces") {
		log.EnableStacks()
	}

	return cfg, nil
}

func init() {
	RootCmd.Flags().StringP("listen", "l", "localhost:8081", "Listen address:port for the application.")

	RootCmd.Flags().Bool("api-only", false, "Starts in API-only mode (schedules & notifications will not be processed). Useful in clusters.")

	RootCmd.Flags().Int("db-max-open", 15, "Max open DB connections.")
	RootCmd.Flags().Int("db-max-idle", 5, "Max idle DB connections.")

	RootCmd.Flags().Int64("max-request-body-bytes", 256*1024, "Max body size for all incoming requests (in bytes). Set to 0 to disable limit.")
	RootCmd.Flags().Int("max-request-header-bytes", 4096, "Max header size for all incoming requests (in bytes). Set to 0 to disable limit.")

	RootCmd.Flags().String("github-base-url", "", "Base URL for GitHub auth and API calls.")
	RootCmd.Flags().String("twilio-base-url", "", "Override the Twilio API URL.")
	RootCmd.Flags().String("slack-base-url", "", "Override the Slack base URL.")

	RootCmd.Flags().String("region-name", "default", "Name of region for message processing (case sensitive). Only one instance per-region-name will process outgoing messages.")

	RootCmd.PersistentFlags().String("db-url", "", "Connection string for Postgres.")
	RootCmd.PersistentFlags().String("db-url-next", "", "Connection string for the *next* Postgres server (enables DB switch-over mode).")

	RootCmd.Flags().String("jaeger-endpoint", "", "Jaeger HTTP Thrift endpoint")
	RootCmd.Flags().String("jaeger-agent-endpoint", "", "Instructs Jaeger exporter to send spans to jaeger-agent at this address.")
	RootCmd.Flags().String("stackdriver-project-id", "", "Project ID for Stackdriver. Enables tracing output to Stackdriver.")
	RootCmd.Flags().String("tracing-cluster-name", "", "Cluster name to use for tracing (i.e. kubernetes, Stackdriver/GKE environment).")
	RootCmd.Flags().String("tracing-pod-namespace", "", "Pod namespace to use for tracing.")
	RootCmd.Flags().String("tracing-pod-name", "", "Pod name to use for tracing.")
	RootCmd.Flags().String("tracing-container-name", "", "Container name to use for tracing.")
	RootCmd.Flags().String("tracing-node-name", "", "Node name to use for tracing.")
	RootCmd.Flags().Float64("tracing-probability", 0.01, "Probability of a new trace to be recorded.")

	RootCmd.Flags().Duration("kubernetes-cooldown", 0, "Cooldown period, from the last TCP connection, before terminating the listener when receiving a shutdown signal.")
	RootCmd.Flags().String("status-addr", "", "Open a port to emit status updates. Connections are closed when the server shuts down. Can be used to keep containers running until GoAlert has exited.")

	RootCmd.PersistentFlags().String("data-encryption-key", "", "Encryption key for sensitive data like signing keys. Used for encrypting new and decrypting existing data.")
	RootCmd.PersistentFlags().String("data-encryption-key-old", "", "Fallback key. Used for decrypting existing data only.")
	RootCmd.PersistentFlags().Bool("stack-traces", false, "Enables stack traces with all error logs.")

	RootCmd.Flags().Bool("stub-notifiers", false, "If true, notification senders will be replaced with a stub notifier that always succeeds (useful for staging/sandbox environments).")

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging.")
	RootCmd.Flags().Bool("log-requests", false, "Log all HTTP requests. If false, requests will be logged for debug/trace contexts only.")
	RootCmd.PersistentFlags().Bool("json", false, "Log in JSON format.")
	RootCmd.PersistentFlags().Bool("log-errors-only", false, "Only log errors (superseeds other flags).")

	RootCmd.Flags().String("ui-url", "", "Proxy UI requests to an alternate host. Default is to serve bundled assets from memory.")
	RootCmd.Flags().Bool("disable-https-redirect", false, "Disable automatic HTTPS redirects.")

	migrateCmd.Flags().String("up", "", "Target UP migration to apply.")
	migrateCmd.Flags().String("down", "", "Target DOWN migration to roll back to.")
	exportCmd.Flags().String("export-dir", "migrations", "Destination dir for export. If it does not exist, it will be created.")

	addUserCmd.Flags().String("user-id", "", "If specified, the auth entry will be created for an existing user ID. Default is to create a new user.")
	addUserCmd.Flags().String("pass", "", "Specify new users password (if blank, prompt will be given).")
	addUserCmd.Flags().String("user", "", "Specifies the login username.")
	addUserCmd.Flags().String("email", "", "Specifies the email address of the new user (ignored if user-id is provided).")
	addUserCmd.Flags().Bool("admin", false, "If specified, the user will be created with the admin role (ignored if user-id is provided).")

	setConfigCmd.Flags().String("data", "", "Use data instead of reading config from stdin.")
	setConfigCmd.Flags().Bool("allow-empty-data-encryption-key", false, "Explicitly allow an empty data-encryption-key when setting config.")

	monitorCmd.Flags().StringP("config-file", "f", "", "Configuration file for monitoring (required).")
	RootCmd.AddCommand(versionCmd, migrateCmd, exportCmd, monitorCmd, switchCmd, addUserCmd, getConfigCmd, setConfigCmd)

	err := viper.BindPFlags(RootCmd.Flags())
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
