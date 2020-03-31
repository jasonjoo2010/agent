package logs

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/projecteru2/agent/engine/logs/transport"
	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
)

// Pool manages and holds a pool of transporters
type Pool struct {
	backends   []*utils.Backend
	workers    *ants.PoolWithFunc
	transports *utils.ObjectPool
	closed     bool
	failed     uint64
	sent       uint64
}

func trasporterFactory(p *Pool) func() interface{} {
	return func() (t interface{}) {
		if p.closed {
			return nil
		}
		b := p.pickServer()
		var err error
		switch b.Type {
		case utils.BlackHole:
			t = transport.NewBlackHole()
		case utils.TCP:
			t, err = transport.NewTCP(b)
		case utils.UDP:
			t, err = transport.NewUDP(b)
		case utils.Journal:
			t, err = transport.NewJournal()
		case utils.Console:
			t = transport.NewConsole()
		case utils.Log:
			t = transport.NewLog()
		default:
			return nil
		}
		if err != nil {
			return nil
		}
		return t
	}
}

func workerFactory(p *Pool) func(interface{}) {
	return func(arg interface{}) {
		logs := arg.([]*types.Log)
		defer ReturnLogsBuffer(logs)
		t := p.fetchOrCreateTrans()
		if t == nil {
			// drop data
			p.increseFailed(len(logs))
			return
		}
		if !t.Send(logs) {
			t.Close()
			return
		}
		p.transports.Put(t)
	}
}

// NewPool creates a pool from specific backends
// return nil and non-empty error when any error occurred
func NewPool(backends []*utils.Backend, maxConcurrency int) (p *Pool, err error) {
	if len(backends) == 0 {
		return nil, errors.New("No log forwarder backends")
	}
	if maxConcurrency < 1 {
		return nil, errors.New("Wrong maxConcurrency for pool")
	}
	// seed once, before transporters are created
	rand.Seed(time.Now().UnixNano())
	p = &Pool{
		backends: backends,
	}
	// init object pool for connections
	p.transports = utils.NewObjectPool(maxConcurrency, trasporterFactory(p), func(obj interface{}) {
		if t, ok := obj.(Transporter); ok {
			t.Close()
		}
	})
	// init workers
	workerPool, err := ants.NewPoolWithFunc(maxConcurrency, workerFactory(p))
	if err != nil {
		return nil, err
	}
	p.workers = workerPool
	return
}

func (p *Pool) fetchOrCreateTrans() Transporter {
	retryCount := 2
	for {
		retryCount--
		if retryCount < 0 {
			// failed
			return nil
		}
		obj := p.transports.Get()
		if obj == nil {
			continue
		}
		t, ok := obj.(Transporter)
		if !ok || t.IsClose() {
			continue
		}
		return t
	}
}

// Failed return the count of failed logs
func (p *Pool) Failed() uint64 {
	return p.failed
}

// ResetFailed reset the failed statistics and return the old value
func (p *Pool) ResetFailed() uint64 {
	val := atomic.SwapUint64(&p.failed, 0)
	return val
}

func (p *Pool) increseFailed(cnt int) {
	atomic.AddUint64(&p.failed, uint64(cnt))
}

// Sent return the count of sent logs
func (p *Pool) Sent() uint64 {
	return p.sent
}

// ResetSent reset the sent statistics and return the old value
func (p *Pool) ResetSent() uint64 {
	val := atomic.SwapUint64(&p.sent, 0)
	return val
}

func (p *Pool) increseSent(cnt int) {
	atomic.AddUint64(&p.sent, uint64(cnt))
}

// pickServer picks a server to connect
func (p *Pool) pickServer() *utils.Backend {
	return p.backends[rand.Intn(len(p.backends))]
}

// Send sends logs out and block if all workers is full
func (p *Pool) Send(logs []*types.Log) bool {
	if p.closed {
		p.increseFailed(len(logs))
		ReturnLogsBuffer(logs)
		return false
	}
	if p.workers.Invoke(logs) != nil {
		p.increseFailed(len(logs))
		return false
	}
	p.increseSent(len(logs))
	return true
}

func (p *Pool) Close() {
	p.closed = true
	p.workers.Release()
	for {
		obj := p.transports.Get()
		if obj == nil {
			break
		}
		t, ok := obj.(Transporter)
		if !ok {
			continue
		}
		t.Close()
	}
}
