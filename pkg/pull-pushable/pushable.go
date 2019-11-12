package pull_pushable

import (
	"github.com/eapache/queue"
	"github.com/sirupsen/logrus"
	pullstream "scuttlebutt-golang/pkg/pull-stream"
	"strconv"
)

var getPushableName = func() func() string {
	counter := 1
	return func() string { return strconv.Itoa(counter) }
}()

var buffer *queue.Queue

type OnClose func(end pullstream.EndOrError)
type End func(end pullstream.EndOrError)
type Push func(data interface{})

func Pushable(name string, onClose OnClose) (pullstream.Read, End, Push) {
	log := logrus.WithField("pushable", name)

	// indicates that the downstream want's to abort the stream
	var abort pullstream.EndOrError
	var ended pullstream.EndOrError

	var cb pullstream.SourceCallback

	//todo. 非线程安全
	buffer = queue.New()

	callback := func(end pullstream.EndOrError, data interface{}) {
		// if error and pushable passed onClose, call it
		// the first time this stream ends or errors.
		if ended.Yes() && (onClose != nil) {
			log.WithField("end", end).Debug("call onClose back with argument")
			onClose(end)
			onClose = nil
		}
		if cb != nil {
			cb(end, data)
			log.WithField("end", end).Debug("call callback with argument")
			//todo. 线程安全?
			cb = nil
		}
	}

	drain := func() {
		if cb == nil {
			return
		}

		if abort.Yes() {
			callback(abort, nil)
		} else if buffer.Length() == 0 && ended.Yes() {
			callback(ended, nil)
		} else if buffer.Length() > 0 {
			callback(pullstream.Null, buffer.Remove())
		}
	}

	end := func(end pullstream.EndOrError) {
		log.Debug("end has been called")
		if !ended.Yes() {
			if end.Yes() {
				ended = end
			} else {
				ended = pullstream.End
			}
		}
		// attempt to drain
		drain()
	}

	push := func(data interface{}) {
		log.WithField("end", ended).WithField("data", data).Debug("push")
		if ended.Yes() {
			return
		}

		// if sink already waiting,
		// we can call back directly.
		if cb != nil {
			callback(abort, data)
			return
		}
		// otherwise, buffer data
		buffer.Add(data)
	}

	read := func(end pullstream.EndOrError, _cb pullstream.SourceCallback) {
		log.WithField("end", end).Debug("read")
		if end.Yes() {
			abort = end
			// if there is already a cb waiting, abort it.
			if cb != nil {
				callback(abort, nil)
			}
		}
		cb = _cb
		drain()
	}

	return read, end, push
}
