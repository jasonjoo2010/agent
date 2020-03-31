package transport

import "github.com/projecteru2/agent/types"

// BlackHole is a black hole which just ignore all logs it receives
type BlackHole struct {
	closed bool
}

func NewBlackHole() *BlackHole {
	return &BlackHole{}
}

func (t *BlackHole) Send(string []*types.Log) bool {
	// Just discard
	return true
}

func (t *BlackHole) IsClose() bool {
	return t.closed
}

func (t *BlackHole) Close() {
	// Nothing needed to be done
	t.closed = true
}
