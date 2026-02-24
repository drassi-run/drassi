package lease

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRenewAt_AlreadyExpired(t *testing.T) {
	d := renewAt(time.Now().Add(-time.Minute))
	assert.LessOrEqual(t, d, time.Duration(0))
}

func TestRenewAt_ShortTime(t *testing.T) {
	// 2 min from now → max(3/4*2m=90s, 2m-1m=1m) = 90m
	d := renewAt(time.Now().Add(2 * time.Minute))
	assert.Less(t, 90*time.Second-d, time.Millisecond)
}

func TestRenewAt_FarFuture(t *testing.T) {
	// 10 min from now → max(3/4*10m=7.5m, 10m-1m=9m) = 9m
	d := renewAt(time.Now().Add(10 * time.Minute))
	assert.Less(t, 9*time.Minute-d, time.Millisecond)
}
