package transport

import (
	"bufio"
	"errors"
	"net"
	"time"

	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
	log "github.com/sirupsen/logrus"
)

// TCP implements transportation through tcp connection
type TCP struct {
	conn   *net.TCPConn
	wr     *bufio.Writer
	closed bool
}

// connectTCP return a connection with the specified address with timeout
func connectTCP(b *utils.Backend) (*net.TCPConn, error) {
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}
	// Timeout should be paid attention during nslookup
	conn, err := dialer.Dial("tcp", b.HostWithPort)
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		// should not happen
		conn.Close()
		return nil, errors.New("Create tcp connection failed")
	}
	return tcpConn, nil
}

func NewTCP(b *utils.Backend) (*TCP, error) {
	if b.Type != utils.TCP {
		return nil, utils.IllegalType
	}
	conn, err := connectTCP(b)
	if err != nil {
		return nil, err
	}
	if conn.SetKeepAlivePeriod(time.Minute) != nil ||
		conn.SetKeepAlive(true) != nil ||
		conn.SetWriteBuffer(1024*1024) != nil ||
		conn.SetNoDelay(true) != nil {
		conn.Close()
		return nil, errors.New("Set options to tcp connection failed")
	}
	return &TCP{
		conn: conn,
		wr:   bufio.NewWriterSize(conn, 128*1024),
	}, nil
}

func (t *TCP) Send(logArr []*types.Log) bool {
	err := transportInJson(t.wr, logArr, -1)
	if err == nil {
		return true
	}

	// io error
	log.Errorf("[TCP] Sending log failed %s", err)
	t.Close()
	return false
}

func (t *TCP) IsClose() bool {
	return t.closed
}

func (t *TCP) Close() {
	t.closed = true
	t.wr.Flush()
	t.conn.Close()
}
