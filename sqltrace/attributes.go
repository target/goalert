package sqltrace

import (
	"net/url"
	"strings"

	"go.opencensus.io/trace"
)

func getConnAttributes(name string) ([]trace.Attribute, error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, err
	}

	return []trace.Attribute{
		trace.StringAttribute("sql.user", u.User.Username()),
		trace.StringAttribute("sql.db", strings.TrimPrefix(u.Path, "/")),
		trace.StringAttribute("sql.host", u.Host),
	}, nil
}
