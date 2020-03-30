package transport

import (
	"github.com/projecteru2/agent/types"
	"github.com/sirupsen/logrus"
)

// Log is a bridged transporter to sirupsen/logrus
type Log struct {
	closed bool
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Send(logs []*types.Log) bool {
	for _, item := range logs {
		// XXX: can be completed later for better formatting
		logrus.Info(item.Data)
	}
	return true
}

func (l *Log) IsClose() bool {
	return l.closed
}

func (l *Log) Close() {
	// Nothing needed to be done
	l.closed = true
}
