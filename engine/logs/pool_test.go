package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
	"github.com/stretchr/testify/assert"
)

func clientRoutine(t *testing.T, client net.Conn, ch chan string) {
	defer client.Close()
	client.SetReadDeadline(time.Now().Add(time.Second * 5))

	rd := bufio.NewReader(client)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			break
		}
		log := &types.Log{}
		err = json.Unmarshal(line, log)
		assert.Nil(t, err)
		assert.Contains(t, log.Name, "test")
		ch <- log.Name
	}
}

func TestPool(t *testing.T) {
	ch := make(chan string)

	// server
	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:14444")
		assert.Nil(t, err)
		//defer l.Close()

		for {
			client, err := l.Accept()
			assert.Nil(t, err)
			go clientRoutine(t, client, ch)
		}
	}()

	backends, err := utils.ParseBackends([]string{
		"tcp://127.0.0.1:14444",
		"tcp://127.0.0.1:14445",
	})
	p, err := NewPool(backends, 3)
	assert.NotNil(t, p)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(p.backends))

	b := p.pickServer()
	assert.NotNil(t, b)
	assert.Equal(t, utils.TCP, b.Type)
	assert.Equal(t, "127.0.0.1", b.Host)

	assert.True(t, p.Send([]*types.Log{
		&types.Log{Name: fmt.Sprint("test", 0)},
		&types.Log{Name: fmt.Sprint("test", 1)},
		&types.Log{Name: fmt.Sprint("test", 2)},
		&types.Log{Name: fmt.Sprint("test", 3)},
		&types.Log{Name: fmt.Sprint("test", 4)},
	}))
	for i := 5; i < 10; i++ {
		p.Send([]*types.Log{
			&types.Log{Name: fmt.Sprint("test", i)},
		})
	}

	for i := 0; i < 10-int(p.Droped()); i++ {
	INNER_LOOP:
		for i < 10-int(p.Droped()) {
			select {
			case name := <-ch:
				assert.Contains(t, name, "test")
				break INNER_LOOP
			case <-time.After(200 * time.Millisecond):
			}
		}
	}
	if p.Droped() > 0 {
		fmt.Println(p.Droped(), "packet(s) droped")
	}

	p.Close()
}
