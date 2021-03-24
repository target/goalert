package app

import (
	"crypto/tls"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
)

type Config struct {
	ListenAddr  string
	Verbose     bool
	JSON        bool
	LogRequests bool
	APIOnly     bool

	TLSListenAddr string
	TLSConfig     *tls.Config

	SysAPIListenAddr string
	SysAPICertFile   string
	SysAPIKeyFile    string

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

	JaegerEndpoint      string
	JaegerAgentEndpoint string

	StackdriverProjectID string

	TracingClusterName   string
	TracingPodNamespace  string
	TracingPodName       string
	TracingContainerName string
	TracingNodeName      string

	KubernetesCooldown time.Duration
	StatusAddr         string

	LogTraces        bool
	TraceProbability float64

	EncryptionKeys keyring.Keys

	RegionName string

	StubNotifiers bool

	UIURL string

	// InitialConfig will be pushed into the config store
	// if specified before the engine is started.
	InitialConfig *config.Config
}
