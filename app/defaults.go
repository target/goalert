package app

// Defaults returns the default app config.
func Defaults() Config {
	return Config{
		DBMaxOpen:         15,
		DBMaxIdle:         5,
		ListenAddr:        "localhost:8081",
		MaxReqBodyBytes:   256 * 1024,
		MaxReqHeaderBytes: 4096,
		RegionName:        "default",
		TraceProbability:  0.01,
	}
}
