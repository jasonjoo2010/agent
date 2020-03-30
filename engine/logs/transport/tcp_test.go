package transport

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
	"github.com/stretchr/testify/assert"
)

func TestTCPConnect(t *testing.T) {
	// not valid ip:port
	conn, err := connectTCP(utils.NewBackend(utils.TCP, "127.0.0.1", 12340))
	assert.Nil(t, conn)
	assert.NotNil(t, err)

	// valid scenario, address may be adjusted anytime to a valid one
	conn, err = connectTCP(utils.NewBackend(utils.TCP, "39.156.69.79", 80))
	assert.NotNil(t, conn)
	assert.Nil(t, err)
	conn.Close()
}

func TestTCPWrongType(t *testing.T) {
	trans, err := NewTCP(&utils.Backend{
		Type: utils.BlackHole,
	})
	assert.Nil(t, trans)
	assert.NotNil(t, err)
	assert.Equal(t, utils.IllegalType, err)
}

func TestTCPNormal(t *testing.T) {
	resultC := make(chan int)

	const (
		total int = 5
		round int = 15
	)
	go func() {
		defer func() {
			resultC <- 1
		}()
		l, err := net.Listen("tcp", "127.0.0.1:14444")
		assert.Nil(t, err)
		defer l.Close()
		client, err := l.Accept()
		assert.Nil(t, err)
		defer client.Close()
		client.SetReadDeadline(time.Now().Add(10 * time.Second))
		rd := bufio.NewReader(client)

		for i := 0; i < total*round; i++ {
			line, _, _ := rd.ReadLine()
			assert.NotNil(t, line)
			assert.NotEmpty(t, line)
			assert.Contains(t, string(line), fmt.Sprint("test", i/round))
		}

		for i := 0; i < total*round; i++ {
			line, _, _ := rd.ReadLine()
			assert.NotNil(t, line)
			assert.NotEmpty(t, line)
			assert.Contains(t, string(line), fmt.Sprint("test", i/round))
		}
	}()

	trans, err := NewTCP(utils.NewBackend(utils.TCP, "127.0.0.1", 14444))
	assert.NotNil(t, trans)
	assert.Nil(t, err)

	data := strings.Repeat("demo", 500)

	for i := 0; i < total; i++ {
		ret := trans.Send([]*types.Log{
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
			&types.Log{Name: fmt.Sprint("test", i)},
		})
		assert.True(t, ret)
		time.Sleep(100 * time.Millisecond)
	}

	for i := 0; i < total; i++ {
		ret := trans.Send([]*types.Log{
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
			&types.Log{Name: fmt.Sprint("test", i), Data: data},
		})
		assert.True(t, ret)
		time.Sleep(100 * time.Millisecond)
	}

	<-resultC

	trans.Close()
	assert.True(t, trans.IsClose())
}
