package logs

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/projecteru2/agent/engine/logs/transport"
	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/utils"
)

// Poll manages and holds a pool of transporters
type Pool struct {
	backends   []*utils.Backend
	workers    *ants.PoolWithFunc
	transports *sync.Pool
	closed     bool
	droped     int32
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
		t := p.fetchOrCreateTrans()
		if t == nil {
			// drop data
			atomic.AddInt32(&p.droped, int32(len(logs)))
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
	// seed once
	rand.Seed(time.Now().UnixNano())
	p = &Pool{
		backends: backends,
	}
	p.transports = &sync.Pool{
		New: trasporterFactory(p),
	}
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

// Droped return the count of droped logs
func (p *Pool) Droped() int32 {
	return p.droped
}

// ResetDroped reset the droped statistics and return the old value
func (p *Pool) ResetDroped() int32 {
	val := atomic.SwapInt32(&p.droped, 0)
	return val
}

// pickServer picks a server to connect
func (p *Pool) pickServer() *utils.Backend {
	return p.backends[rand.Intn(len(p.backends))]
}

// Send sends logs out and block if all workers is full
func (p *Pool) Send(logs []*types.Log) bool {
	if p.closed {
		return false
	}
	p.workers.Invoke(logs)
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
