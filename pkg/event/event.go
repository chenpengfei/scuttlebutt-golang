package event

import "sync"

type OnCallback func(data interface{})

//todo. did need ?
var rw sync.RWMutex

type Event struct {
	store map[string][]OnCallback
}

func NewEvent() *Event {
	return &Event{
		store: make(map[string][]OnCallback),
	}
}

func (e *Event) Emit(name string, data interface{}) {
	rw.RLock()
	if _, ok := e.store[name]; ok {
		for _, cb := range e.store[name] {
			cb(data)
		}
	}
	rw.RUnlock()
}

func (e *Event) On(name string, cb OnCallback) {
	rw.Lock()
	if _, ok := e.store[name]; !ok {
		e.store[name] = make([]OnCallback, 0)
	}
	e.store[name] = append(e.store[name], cb)
	rw.Unlock()
}

func (e *Event) Once(name string, data interface{}) {
	//todo
}

func (e *Event) RemoveListener(name string, cb OnCallback) {
	//if _, found := e.store[name]; found {
		//todo
	//}
}
