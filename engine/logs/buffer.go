package logs

import (
	"sync"

	"github.com/projecteru2/agent/types"
)

const BATCH_SIZE int = 400

var logsBufferPool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return make([]*types.Log, 0, BATCH_SIZE)
	},
}

// GetLogsBuffer trys to get or create a new buffer from pool
func GetLogsBuffer() []*types.Log {
	obj := logsBufferPool.Get()
	if obj == nil {
		return nil
	}
	buf, ok := obj.([]*types.Log)
	if !ok {
		// should not happen
		return nil
	}
	return buf
}

// ReturnLogsBuffer receives the useless buffer for next time usage
func ReturnLogsBuffer(buffer []*types.Log) {
	buffer = buffer[:0]
	logsBufferPool.Put(buffer)
}
