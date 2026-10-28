package twilio

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestWrapStatusErr(t *testing.T) {
	notFound := &Exception{Status: 404, Code: 20404, Message: "The requested resource was not found"}
	err := wrapStatusErr(notFound)
	assert.ErrorIs(t, err, notification.ErrStatusNotFound)

	var e *Exception
	assert.True(t, errors.As(err, &e), "original exception should be preserved")
	assert.Equal(t, 20404, e.Code)

	other := &Exception{Status: 500, Code: 20500, Message: "Internal Server Error"}
	assert.NotErrorIs(t, wrapStatusErr(other), notification.ErrStatusNotFound)

	plain := errors.New("connection reset")
	assert.Equal(t, plain, wrapStatusErr(plain))
}
