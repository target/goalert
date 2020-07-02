package app

import (
	"crypto/tls"
	"time"

	"github.com/target/goalert/keyring"
)

type appConfig struct {
	ListenAddr  string
	Verbose     bool
	JSON        bool
	LogRequests bool
	APIOnly     bool

	TLSListenAddr string
	TLSConfig     *tls.Config

	HTTPPrefix string

	DBMaxOpen int
	DBMaxIdle int

	MaxReqBodyBytes   int64
	MaxReqHeaderBytes int

	DisableHTTPSRedirect bool

	TwilioBaseURL   string
	TwilioLookupURL string
	SlackBaseURL    string

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
}
