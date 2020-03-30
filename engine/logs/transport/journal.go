package transport

import (
	"errors"
	"fmt"

	"github.com/coreos/go-systemd/journal"
	"github.com/projecteru2/agent/types"
)

// Journal implements transportation to remote journal deamon
// Because the imeplementation always shares the same unixgram socket
// so there's no need to use multiple Journals
type Journal struct {
	closed bool
}

var JournalDisabled = errors.New("journal disabled")

func NewJournal() (*Journal, error) {
	if !journal.Enabled() {
		return nil, JournalDisabled
	}
	return &Journal{}, nil
}

func sendLog(logline *types.Log) bool {
	vars := map[string]string{
		"SYSLOG_IDENTIFIER": logline.Name,
		"ID":                logline.ID,
		"TYPE":              logline.Type,
		"ENTRY_POINT":       logline.EntryPoint,
		"IDENT":             logline.Ident,
		"DATE_TIME":         logline.Datetime,
		"EXTRA":             fmt.Sprintf("%v", logline.Extra),
	}

	p := fmt.Sprintf("message %s", logline.Data)
	return journal.Send(p, journal.PriErr, vars) == nil
}

func (j *Journal) Send(logArr []*types.Log) bool {
	for _, logline := range logArr {
		sendLog(logline)
	}
	return true
}

func (j *Journal) IsClose() bool {
	return j.closed
}

func (j *Journal) Close() {
	j.closed = true
}
