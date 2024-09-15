package xdig

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func newScope() *dig.Scope {
	c := dig.New()
	return c.Scope("test")
}

func TestDig(t *testing.T) {
	scope := newScope()

	// Test Supply
	err := Supply(scope, "foobar")
	assert.NoError(t, err)
	err = scope.Invoke(func(s string) {
		assert.Equal(t, "foobar", s)
	})
	assert.NoError(t, err)

	// Test Populate
	var str string
	err = Populate(scope, &str)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", str)

	// Test Replace
	err = Replace(scope, "abcxyz")
	assert.NoError(t, err)
	err = scope.Invoke(func(s string) {
		assert.Equal(t, "abcxyz", s)
	})
	assert.NoError(t, err)
}
