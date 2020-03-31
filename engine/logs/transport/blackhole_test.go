package transport

import (
	"testing"

	"github.com/projecteru2/agent/types"
	"github.com/stretchr/testify/assert"
)

func TestBlackHole(t *testing.T) {
	trans := NewBlackHole()
	assert.NotNil(t, trans)
	assert.False(t, trans.IsClose())
	assert.True(t, trans.Send(nil))
	assert.True(t, trans.Send([]*types.Log{}))
	assert.True(t, trans.Send([]*types.Log{
		&types.Log{},
		&types.Log{},
	}))

	trans.Close()
	assert.True(t, trans.IsClose())
}
