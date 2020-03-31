package logs

import (
	"errors"
	"sync/atomic"

	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var (
	BufferFull    = errors.New("Sending queue of logs is full, discard")
	AlreadyClosed = errors.New("Writer has been closed")
	FlowLimiting  = errors.New("Flow limiting")
)

// Writer is a writer!
type Writer struct {
	pool    *Pool
	buf     chan *types.Log
	limiter *rate.Limiter
	dropped uint64
	closed  bool
}

// NewWriter create a writer based on specified backends
//	concurrency specify the parellel workers' count
//	bufferSize specify the queue capacity which are waiting to be processed
//	ratelimit limit how many logs can be sent out per second, -1 for no limit
func NewWriter(backends []*utils.Backend, concurrency, bufferSize, ratelimit int) (*Writer, error) {
	var (
		p   *Pool
		err error
	)
	if len(backends) < 1 {
		p, err = NewPool([]*utils.Backend{
			utils.NewBackend(utils.BlackHole, "", 0),
		}, 1)
	} else {
		p, err = NewPool(backends, concurrency)
	}
	if err != nil {
		return nil, err
	}
	w := &Writer{
		pool: p,
		buf:  make(chan *types.Log, bufferSize),
	}
	if ratelimit > 0 {
		w.limiter = rate.NewLimiter(rate.Limit(ratelimit), ratelimit/10)
	}
	go w.drainer()
	return w, nil
}

// drainer is the worker
func (w *Writer) drainer() {
	for !w.closed {
		arr := GetLogsBuffer()
	FETCH_LOOP:
		for i := 0; i < cap(arr); i++ {
			select {
			case log := <-w.buf:
				arr = append(arr, log)
			default:
				break FETCH_LOOP
			}
		}
		if len(arr) == 0 {
			// empty
			//time.Sleep(time.Millisecond * 50)
			continue
		}
		if !w.pool.Send(arr) {
			// failed, do nothing currently in writer
		}
	}
	logrus.Info("Drainer of writer stopped")
}

// Droped return the count of dropped logs
func (w *Writer) Dropped() uint64 {
	return w.dropped
}

// ResetDroped reset the droped statistics and return the old value
func (w *Writer) ResetDropped() uint64 {
	val := atomic.SwapUint64(&w.dropped, 0)
	return val
}

func (w *Writer) increaseDropped() uint64 {
	val := atomic.AddUint64(&w.dropped, 1)
	return val
}

func (w *Writer) Failed() uint64 {
	return w.pool.Failed()
}

func (w *Writer) ResetFailed() uint64 {
	return w.pool.ResetFailed()
}

func (w *Writer) Sent() uint64 {
	return w.pool.Sent()
}

func (w *Writer) ResetSent() uint64 {
	return w.pool.ResetSent()
}

func (w *Writer) Close() {
	w.closed = true
	w.pool.Close()
	close(w.buf)
}

// Write write log to remote
func (w *Writer) Write(logline *types.Log) error {
	if w.closed {
		w.increaseDropped()
		return AlreadyClosed
	}
	if w.limiter != nil && !w.limiter.Allow() {
		// limited
		w.increaseDropped()
		return FlowLimiting
	}
	select {
	case w.buf <- logline:
		// succ
		return nil
	default:
		w.increaseDropped()
		return BufferFull
	}
}
