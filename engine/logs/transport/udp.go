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

// TCP implements transportation through udp connection
type UDP struct {
	conn   *net.UDPConn
	wr     *bufio.Writer
	closed bool
}

// connectUDP return a connection with the specified address with timeout
func connectUDP(b *utils.Backend) (*net.UDPConn, error) {
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}
	// Timeout should be paid attention during nslookup
	conn, err := dialer.Dial("udp", b.HostWithPort)
	if err != nil {
		return nil, err
	}
	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		// should not happen
		conn.Close()
		return nil, errors.New("Create udp connection failed")
	}
	return udpConn, nil
}

func NewUDP(b *utils.Backend) (*UDP, error) {
	if b.Type != utils.UDP {
		return nil, utils.IllegalType
	}
	conn, err := connectUDP(b)
	if err != nil {
		return nil, err
	}
	if conn.SetWriteBuffer(1024*1024) != nil {
		conn.Close()
		return nil, errors.New("Set options to udp connection failed")
	}
	return &UDP{
		conn: conn,
		wr:   bufio.NewWriterSize(conn, 128*1024),
	}, nil
}

func (u *UDP) Send(logArr []*types.Log) bool {
	err := transportInJson(u.wr, logArr, 1420)
	if err == nil {
		return true
	}

	// io error
	log.Errorf("[UDP] Sending log failed %s", err)
	u.Close()
	return false
}

func (u *UDP) IsClose() bool {
	return u.closed
}

func (u *UDP) Close() {
	u.closed = true
	u.wr.Flush()
	u.conn.Close()
}
