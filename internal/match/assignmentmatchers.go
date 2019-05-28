package match

import (
	"fmt"
	"github.com/target/goalert/assignment"

	"github.com/golang/mock/gomock"
)

func Target(t assignment.Target) gomock.Matcher {
	return asnTgtMatcher{Target: t}
}
func TargetValue(t assignment.TargetType, id string) gomock.Matcher {
	return Target(assignment.RawTarget{
		ID:   id,
		Type: t,
	})
}
func Source(s assignment.Source) gomock.Matcher {
	return asnSrcMatcher{Source: s}
}
func SourceValue(s assignment.SrcType, id string) gomock.Matcher {
	return Source(assignment.RawSource{
		ID:   id,
		Type: s,
	})
}

type asnTgtMatcher struct{ assignment.Target }

func (m asnTgtMatcher) Matches(x interface{}) bool {
	t, ok := x.(assignment.Target)
	if !ok {
		return false
	}
	return t.TargetType() == m.TargetType() && t.TargetID() == m.TargetID()
}
func (m asnTgtMatcher) String() string {
	return fmt.Sprintf("%s(%s)", m.TargetType().String(), m.TargetID())
}

type asnSrcMatcher struct{ assignment.Source }

func (m asnSrcMatcher) Matches(x interface{}) bool {
	s, ok := x.(assignment.Source)
	if !ok {
		return false
	}
	return s.SourceType() == m.SourceType() && s.SourceID() == m.SourceID()
}
func (m asnSrcMatcher) String() string {
	return fmt.Sprintf("%s(%s)", m.SourceType().String(), m.SourceID())
}
