package graphql2

import (
	"github.com/target/goalert/schedule/rule"
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalClockTime(c rule.Clock) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+c.String()+"\"")
	})
}
func UnmarshalClockTime(v interface{}) (rule.Clock, error) {
	str, ok := v.(string)
	if !ok {
		return rule.Clock(0), errors.New("ClockTime must be strings")
	}
	return rule.ParseClock(str)
}
