package logs

import (
	"github.com/projecteru2/agent/types"
)

// Transporter defines general transporting abstraction and it can be pooled(Multiple instances).
// According to this it's designed as NONE THREAD-SAFE.
type Transporter interface {
	// Send send out data in array and return true when success false for failure
	Send(string []*types.Log) bool

	// Close the transporter and it should never be used again
	Close()

	// IsClose whether it's a closed transporter
	IsClose() bool
}
