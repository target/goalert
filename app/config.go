package app

import (
	"crypto/tls"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/util/log"
)

type Config struct {
	Logger *log.Logger

	ListenAddr  string
	Verbose     bool
	JSON        bool
	LogRequests bool
	APIOnly     bool
	LogEngine   bool

	PublicURL string

	TLSListenAddr string
	TLSConfig     *tls.Config

	SysAPIListenAddr string
	SysAPICertFile   string
	SysAPIKeyFile    string
	SysAPICAFile     string

	HTTPPrefix string

	DBMaxOpen int
	DBMaxIdle int

	MaxReqBodyBytes   int64
	MaxReqHeaderBytes int

	DisableHTTPSRedirect bool

	TwilioBaseURL string
	SlackBaseURL  string

	DBURL     string
	DBURLNext string

	KubernetesCooldown time.Duration
	StatusAddr         string

	EngineCycleTime time.Duration

	EncryptionKeys keyring.Keys

	RegionName string

	StubNotifiers bool

	UIDir string

	// InitialConfig will be pushed into the config store
	// if specified before the engine is started.
	InitialConfig *config.Config
}
