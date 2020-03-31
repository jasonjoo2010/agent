package transport

import (
	"bufio"
	"fmt"
	"io"
	"testing"

	"github.com/projecteru2/agent/types"
	"github.com/stretchr/testify/assert"
)

func TestTransportInJson(t *testing.T) {
	notifier := make(chan int)
	const total int = 100
	r, w := io.Pipe()
	rd, wr := bufio.NewReader(r), bufio.NewWriter(w)
	go func() {
		defer func() {
			notifier <- 1
		}()
		for i := 0; i < total; i++ {
			line, _, _ := rd.ReadLine()
			assert.Contains(t, string(line), fmt.Sprint("test", i))
		}
	}()

	for i := 0; i < total; i++ {
		transportInJson(wr, []*types.Log{
			&types.Log{
				Name: fmt.Sprint("test", i),
			},
		}, -1)
	}

	<-notifier
}
