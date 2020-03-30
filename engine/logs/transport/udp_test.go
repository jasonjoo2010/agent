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

func TestUDPConnect(t *testing.T) {
	// not valid ip:port, but it still can succeed
	conn, err := connectUDP(utils.NewBackend(utils.UDP, "127.0.0.1", 12340))
	assert.NotNil(t, conn)
	assert.Nil(t, err)
	conn.Close()

	// valid scenario, address may be adjusted anytime to a valid one
	conn, err = connectUDP(utils.NewBackend(utils.UDP, "114.114.114.114", 53))
	assert.NotNil(t, conn)
	assert.Nil(t, err)
	conn.Close()
}

func TestUDPWrongType(t *testing.T) {
	trans, err := NewUDP(&utils.Backend{
		Type: utils.BlackHole,
	})
	assert.Nil(t, trans)
	assert.NotNil(t, err)
	assert.Equal(t, utils.IllegalType, err)
}

func TestUDPNormal(t *testing.T) {
	resultC := make(chan int)

	// This test is not strict because UDP doesn't guarantee the orders
	const (
		total int = 5
		round int = 15
	)
	go func() {
		defer func() {
			resultC <- 1
		}()
		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:14444")
		assert.NotNil(t, addr)
		assert.Nil(t, err)
		l, err := net.ListenUDP("udp", addr)
		assert.Nil(t, err)
		defer l.Close()
		l.SetReadDeadline(time.Now().Add(10 * time.Second))
		rd := bufio.NewReaderSize(l, 1024*1024)
		// small logs
		for i := 0; i < total*round; i++ {
			line, _, _ := rd.ReadLine()
			assert.NotNil(t, line)
			assert.NotEmpty(t, line)
			assert.Contains(t, string(line), fmt.Sprint("test", i/round))
			println(string(line))
		}

		// large logs
		for i := 0; i < total*round; i++ {
			line, _, _ := rd.ReadLine()
			assert.NotNil(t, line)
			assert.NotEmpty(t, line)
			assert.Contains(t, string(line), fmt.Sprint("test", i/round))
		}
	}()

	trans, err := NewUDP(utils.NewBackend(utils.UDP, "127.0.0.1", 14444))
	assert.NotNil(t, trans)
	assert.Nil(t, err)

	data := strings.Repeat("demo", 500)

	// small logs
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

	// large logs
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
