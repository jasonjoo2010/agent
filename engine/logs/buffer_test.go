package logs

import (
	"fmt"
	"testing"

	"github.com/projecteru2/agent/types"
	"github.com/stretchr/testify/assert"
)

func TestLogsBuffer(t *testing.T) {
	buf := GetLogsBuffer()
	assert.NotNil(t, buf)
	assert.Equal(t, 0, len(buf))
	assert.Equal(t, 100, cap(buf))
	buf = append(buf, &types.Log{})
	assert.Equal(t, 1, len(buf))
	assert.Equal(t, 100, cap(buf))

	ReturnLogsBuffer(buf)

	buf1 := GetLogsBuffer()
	assert.Equal(t, fmt.Sprintf("%p", buf), fmt.Sprintf("%p", buf1))
	assert.Equal(t, 0, len(buf1))
	assert.Equal(t, 100, cap(buf1))
}
