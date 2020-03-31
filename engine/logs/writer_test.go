package logs

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
)

func clientReading(client *net.TCPConn, ch chan int) {
	defer client.Close()
	rd := bufio.NewReader(client)

	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "test") {
			ch <- 1
		}
	}
}

func TestNewWriterWithTCP(t *testing.T) {
	counter := make(chan int)

	go func() {
		tcpL, err := net.Listen("tcp", "127.0.0.1:34567")
		assert.NoError(t, err)
		defer tcpL.Close()

		for {
			client, err := tcpL.Accept()
			if err == nil {
				go clientReading(client.(*net.TCPConn), counter)
			}
		}
	}()

	// tcp writer
	addr := "tcp://127.0.0.1:34567"
	backend, _ := utils.ParseBackend(addr)
	w, err := NewWriter([]*utils.Backend{backend}, 2, 1000, -1)
	assert.NoError(t, err)
	const NUM int = 10
	go func() {
		for i := 0; i < NUM; i++ {
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
		}
	}()

	for i := 0; i < NUM*4; i++ {
		<-counter
	}
	w.Close()
}

func benchmark(t *testing.T, ratelimit, port int) {
	counter := make(chan int, 100000)

	go func() {
		tcpL, err := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
		assert.NoError(t, err)
		defer tcpL.Close()

		for {
			client, err := tcpL.Accept()
			if err == nil {
				go clientReading(client.(*net.TCPConn), counter)
			}
		}
	}()

	// tcp writer
	addr := fmt.Sprint("tcp://127.0.0.1:", port)
	backend, _ := utils.ParseBackend(addr)
	w, err := NewWriter([]*utils.Backend{backend}, 10, 100000, ratelimit)
	assert.NoError(t, err)
	const NUM int = 1000000
	go func() {
		for i := 0; i < NUM; i++ {
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			w.Write(&types.Log{
				Data: fmt.Sprint("test", i),
			})
			if i%100000 == 0 {
				fmt.Println("send:", w.Sent(), ", dropped:", w.Dropped(), ", failed:", w.Failed())
			}
		}
	}()

	begin := time.Now()
	total := uint64(NUM * 4)
	for i := uint64(0); i < total-w.Dropped()-w.Failed(); i++ {
	LOOP:
		for i < total-w.Dropped()-w.Failed() {
			select {
			case <-counter:
				break LOOP
			default:
				// retry
				time.Sleep(time.Millisecond)
			}
		}
	}
	cost := time.Now().Sub(begin).Seconds()
	fmt.Printf("%.2fs\n", cost)
	fmt.Println("Rate:", int(float64(w.Sent())/cost), "records/s")
	w.Close()
}

func TestNewWriterBenchmarkLocally(t *testing.T) {
	benchmark(t, -1, 34568)
}

func TestNewWriterBenchmarkWithLimiting(t *testing.T) {
	benchmark(t, 2000, 34569)
}
