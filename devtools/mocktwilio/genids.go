package mocktwilio

import (
	"fmt"
	"sync/atomic"
)

var (
	msgSvcID uint64
	phoneNum uint64
)

// NewMsgServiceID is a convenience method that returns a new unique messaging service ID.
func NewMsgServiceID() string { return fmt.Sprintf("MG%032d", atomic.AddUint64(&msgSvcID, 1)) }

// NewPhoneNumber is a convenience method that returns a new unique phone number.
func NewPhoneNumber() string {
	id := atomic.AddUint64(&phoneNum, 1)

	return fmt.Sprintf("+1%d555%04d", 201+id/10000, id%10000)
}
