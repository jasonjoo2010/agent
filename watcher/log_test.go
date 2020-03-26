package watcher

import (
	"bufio"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/projecteru2/agent/common"
	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/core/utils"
	"github.com/stretchr/testify/assert"
)

func mockConsumer(app string) (*types.LogConsumer, net.Conn) {
	server, client := net.Pipe()
	return &types.LogConsumer{
		ID:   utils.RandomString(8),
		App:  app,
		Conn: server,
		Buf:  bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)),
	}, client
}

func TestLifeCycle(t *testing.T) {
	w := GetInstance()
	w1 := GetInstance()
	assert.Equal(t, w, w1)

	// register consumer
	consumer, client := mockConsumer("test")
	reader := bufio.NewReader(client)
	w.ConsumerC <- consumer
	assert.Equal(t, 1, len(w.consumer))
	assert.NotNil(t, w.consumer["test"])
	assert.Equal(t, 1, len(w.consumer["test"]))

	// new log
	w.LogC <- &types.Log{
		ID:         "container1",
		Name:       "test",
		Type:       "t",
		EntryPoint: "entrypoint1",
		Ident:      "a_b_c-d",
		Data:       "Hello",
		Datetime:   time.Now().Format(common.DateTimeFormat),
	}
	line, _, _ := reader.ReadLine()
	l, _ := strconv.ParseInt(string(line), 16, 32)
	assert.True(t, l > 0)
	line, _, _ = reader.ReadLine()
	assert.Contains(t, string(line), "container1")
	assert.Contains(t, string(line), "entrypoint1")
	assert.Contains(t, string(line), "test")
	assert.Contains(t, string(line), "Hello")

	// close
	client.Close()
	consumer.Conn.Close()
	assert.Equal(t, 1, len(w.consumer))
	w.LogC <- &types.Log{
		ID:         "container1",
		Name:       "test",
		Type:       "t",
		EntryPoint: "entrypoint1",
		Ident:      "a_b_c-d",
		Data:       "World",
		Datetime:   time.Now().Format(common.DateTimeFormat),
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(w.consumer))

	assert.True(t, w.Stop())
}
