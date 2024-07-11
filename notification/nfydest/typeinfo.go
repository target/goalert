package nfydest

type TypeInfo struct {
	Type string

	Name        string
	IconURL     string
	IconAltText string
	Enabled     bool

	RequiredFields []FieldConfig
	DynamicParams  []DynamicParamConfig

	UserDisclaimer string

	// Message type info
	SupportsStatusUpdates      bool
	SupportsAlertNotifications bool
	SupportsUserVerification   bool
	SupportsOnCallNotify       bool
	SupportsSignals            bool

	UserVerificationRequired bool
	StatusUpdatesRequired    bool
}

func (t TypeInfo) IsContactMethod() bool {
	return t.SupportsAlertNotifications && t.SupportsUserVerification
}

func (t TypeInfo) IsEPTarget() bool {
	return t.SupportsAlertNotifications && !t.UserVerificationRequired
}

func (t TypeInfo) IsSchedOnCallNotify() bool {
	return t.SupportsOnCallNotify && !t.UserVerificationRequired
}

func (t TypeInfo) IsDynamicAction() bool {
	return t.SupportsSignals && !t.UserVerificationRequired
}

type FieldConfig struct {
	FieldID            string
	Label              string
	Hint               string
	HintURL            string
	PlaceholderText    string
	Prefix             string
	InputType          string
	SupportsSearch     bool
	SupportsValidation bool
}

type DynamicParamConfig struct {
	ParamID      string
	Label        string
	Hint         string
	HintURL      string
	DefaultValue string
}
