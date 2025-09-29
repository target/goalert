package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCycleMonitor_GC(t *testing.T) {
	cm := newCycleMonitor()
	for range 1000 {
		cm.startNextCycle()()
	}

	assert.Len(t, cm.cycles, cycleHist)
}
