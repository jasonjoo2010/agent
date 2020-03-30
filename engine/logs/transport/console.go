package transport

import (
	"fmt"

	"github.com/projecteru2/agent/types"
)

// Console is a simple transporter which just print logs into fd(1)
type Console struct {
	closed bool
}

func NewConsole() *Console {
	return &Console{}
}

func (c *Console) Send(logs []*types.Log) bool {
	for _, item := range logs {
		// XXX: can be completed later for better formatting
		fmt.Printf("[%s] %s\n", item.Datetime, item.Data)
	}
	return true
}

func (c *Console) IsClose() bool {
	return c.closed
}

func (c *Console) Close() {
	// Nothing needed to be done
	c.closed = true
}
