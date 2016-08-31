package status

import (
	"encoding/json"
	"io"
	"sync"

	eventtypes "github.com/docker/engine-api/types/events"
)

type EventHandler struct {
	sync.Mutex
	handlers map[string]func(eventtypes.Message)
}

func NewEventHandler() *EventHandler {
	return &EventHandler{handlers: make(map[string]func(eventtypes.Message))}
}

func (e *EventHandler) Handle(action string, h func(eventtypes.Message)) {
	e.Lock()
	e.handlers[action] = h
	e.Unlock()
}

func (e *EventHandler) Watch(c <-chan eventtypes.Message) {
	for ev := range c {
		//		log.Debugf("cid %s action %s", ev.ID[:7], ev.Action)
		e.Lock()
		h, exists := e.handlers[ev.Action]
		e.Unlock()
		if !exists {
			continue
		}
		go h(ev)
	}
}

func DecodeEvents(input io.Reader, c chan eventtypes.Message) error {
	dec := json.NewDecoder(input)
	for {
		var event eventtypes.Message
		if err := dec.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		c <- event
	}
	return nil
}