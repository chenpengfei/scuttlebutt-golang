package pull_pushable

import (
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/pull-stream/pkg/util"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	"strconv"
)

type OnClose func(end error)
type Ender func(end error)
type Push func(data interface{})

var getPushableName = func() func() string {
	counter := 0
	return func() string {
		counter++
		return strconv.Itoa(counter)
	}
}()

type Pushable struct {
	Read  pull.Read
	Ender Ender
	Push  Push

	name    string
	onClose OnClose
}

func NewPushable(opts ...Option) *Pushable {
	pushable := &Pushable{}

	for _, opt := range opts {
		opt(pushable)
	}

	if pushable.name == "" {
		pushable.name = getPushableName()
	}

	log := logger.WithNamespace(pushable.name)

	// indicates that the downstream want's to abort the stream
	var abort error
	var ended error
	var cb pull.SourceCallback
	buffer := make([]interface{}, 0)

	callback := func(end error, data interface{}) {
		// if error and pushable passed onClose, call it
		// the first time this stream ends or errors.
		if ended != nil && pushable.onClose != nil {
			c := pushable.onClose
			pushable.onClose = nil
			log.WithField("end", end).Debug("call onClose back with argument")
			c(end)
		}

		_cb := cb
		cb = nil
		if _cb != nil {
			log.WithField("end", end).Debug("call callback with argument")
			_cb(end, data)
		}
	}

	drain := func() {
		if cb == nil {
			return
		}
		if abort != nil {
			callback(abort, nil)
		} else if len(buffer) == 0 && ended != nil {
			callback(ended, nil)
		} else if len(buffer) > 0 {
			data := buffer[0]
			buffer = buffer[1:]
			callback(nil, data)
		}
	}

	pushable.Ender = func(end error) {
		log.WithError(end).Debug("end has been called")
		ended = util.ErrOrEnd(util.ErrOr(ended, end))
		// attempt to drain
		drain()
	}

	pushable.Push = func(data interface{}) {
		log.WithError(ended).WithField("data", data).Debug("push a new data")
		if ended != nil {
			return
		}
		// if sink already waiting,
		// we can call back directly.
		if cb != nil {
			callback(abort, data)
			return
		}
		// otherwise, buffer data
		buffer = append(buffer, data)
	}

	pushable.Read = func(end error, _cb pull.SourceCallback) {
		log.WithField("end", end).Debug("read")
		if end != nil {
			abort = end
			// if there is already a cb waiting, abort it.
			if cb != nil {
				callback(abort, nil)
			}
		}
		cb = _cb
		drain()
	}

	return pushable
}
