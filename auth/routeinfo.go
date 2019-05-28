package auth

// RouteInfo represents path information for the current request.
type RouteInfo struct {
	// Relative provides a path, relative to the base of the current
	// identity provider.
	RelativePath string

	// CurrentURL is calculated using the AuthRefererURLs and
	// the current auth attempt's referer. It does not include
	// query parameters of the current request.
	CurrentURL string
}
