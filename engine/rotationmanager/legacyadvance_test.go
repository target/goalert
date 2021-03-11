package rotationmanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/schedule/rotation"
)

func TestCalcOldEndTime(t *testing.T) {
	tzChicago, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	rot := &rotation.Rotation{
		Type:        rotation.TypeHourly,
		ShiftLength: 8,
		Start:       time.Date(2018, 11, 21, 16, 0, 0, 0, time.UTC).In(tzChicago),
	}

	t.Log("Start      =", rot.Start.String())

	shiftStart := time.Date(2021, 3, 10, 15, 0, 1, 0, time.UTC).In(tzChicago)
	t.Log("ShiftStart =", shiftStart.String())

	// New code should ensure handoff @ 10am, 6pm, and 2am
	assert.Equal(t, time.Date(2021, 3, 10, 10, 0, 0, 0, tzChicago).String(), rot.EndTime(shiftStart).String())

	// Old code didn't always do this
	assert.Equal(t, time.Date(2021, 3, 10, 17, 0, 0, 0, tzChicago).String(), calcOldEndTime(rot, shiftStart).String())
}
