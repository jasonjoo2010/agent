package watcher

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/projecteru2/agent/types"
)

// Watcher indicate watcher
type Watcher struct {
	LogC      chan *types.Log
	ConsumerC chan *types.LogConsumer
	consumer  map[string]map[string]*types.LogConsumer
	needStop  bool
	stopped   bool
}

var (
	lock      sync.Mutex
	singleton *Watcher
)

// createInstance create a new watcher internally and begin to serve
func create() *Watcher {
	w := &Watcher{}
	w.consumer = map[string]map[string]*types.LogConsumer{}
	w.LogC = make(chan *types.Log)
	w.ConsumerC = make(chan *types.LogConsumer)
	go w.Serve()
	return w
}

// GetInstance returns the singleton of Watcher
func GetInstance() *Watcher {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleton == nil {
			singleton = create()
		}
	}
	return singleton
}

func (w *Watcher) log(log *types.Log) {
	consumers, ok := w.consumer[log.Name]
	if !ok {
		return
	}
	data, err := json.Marshal(log)
	if err != nil {
		logrus.Error(err)
		return
	}
	line := fmt.Sprintf("%X\r\n%s\r\n\r\n", len(data)+2, string(data))
	for id, consumer := range consumers {
		_, err := consumer.Buf.WriteString(line)
		if err == nil {
			err = consumer.Buf.Flush()
		}
		if err != nil {
			logrus.Error(err)
			logrus.Infof("%s %s log detached", consumer.App, consumer.ID)
			consumer.Conn.Close()
			w.unregisterConsumer(log.Name, id)
		}
	}
}

func (w *Watcher) unregisterConsumer(app, consumerId string) {
	consumers, ok := w.consumer[app]
	if !ok {
		return
	}
	delete(consumers, consumerId)
	if len(w.consumer[app]) == 0 {
		delete(w.consumer, app)
	}
}

func (w *Watcher) registerConsumer(consumer *types.LogConsumer) {
	consumers, ok := w.consumer[consumer.App]
	if !ok {
		w.consumer[consumer.App] = map[string]*types.LogConsumer{}
		consumers = w.consumer[consumer.App]
	}
	consumers[consumer.ID] = consumer
}

// Serve start monitor
func (w *Watcher) Serve() {
	logrus.Info("[logServe] Log monitor started")
	const timeout = 300 * time.Millisecond
	timer := time.NewTimer(timeout)
	for !w.needStop {
		timer.Reset(timeout)
		select {
		case log := <-w.LogC:
			w.log(log)
		case consumer := <-w.ConsumerC:
			w.registerConsumer(consumer)
		case <-timer.C:
			// timeout
		}
	}
	w.stopped = true
	w.needStop = false
}

// Stop will request to termiate the serving progress and wait it to stop
func (w *Watcher) Stop() bool {
	w.needStop = true
	for cycles := 200; cycles > 0; cycles-- {
		if w.stopped {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(w.LogC)
	close(w.ConsumerC)
	return false
}
