package transport

import (
	"testing"
	"time"

	"github.com/projecteru2/agent/types"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	trans := NewLog()
	assert.NotNil(t, trans)
	assert.False(t, trans.IsClose())
	assert.True(t, trans.Send(nil))
	assert.True(t, trans.Send([]*types.Log{}))
	assert.True(t, trans.Send([]*types.Log{
		&types.Log{
			Datetime: time.Now().Local().Format("2006-01-02 15:04:05"),
			Data:     "log1",
		},
		&types.Log{
			Datetime: time.Now().Local().Format("2006-01-02 15:04:05"),
			Data:     "log2",
		},
	}))

	trans.Close()
	assert.True(t, trans.IsClose())
}
