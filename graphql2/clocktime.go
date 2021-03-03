package graphql2

import (
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/timeutil"
)

func MarshalClockTime(c timeutil.Clock) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+c.String()+"\"")
	})
}
func UnmarshalClockTime(v interface{}) (timeutil.Clock, error) {
	str, ok := v.(string)
	if !ok {
		return timeutil.Clock(0), errors.New("ClockTime must be strings")
	}
	return timeutil.ParseClock(str)
}
