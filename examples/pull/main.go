package main

import (
	"github.com/eapache/queue"
	log "github.com/sirupsen/logrus"
	pullpushable "scuttlebutt-golang/pkg/pull-pushable"
	pullstream "scuttlebutt-golang/pkg/pull-stream"
	"strings"
	"time"
)

func sourceQueue(queue *queue.Queue) pullstream.Read {
	return func(end pullstream.EndOrError, cb pullstream.SourceCallback) {
		if end.Yes() {
			return
		}
		if queue.Length() == 0 {
			cb(pullstream.End, nil)
		} else {
			cb(pullstream.Null, queue.Remove())
		}
	}
}

func sink(cb pullstream.SourceCallback) pullstream.Sink {
	return func(read pullstream.Read) {
		go func() {
			for {
				read(pullstream.Null, cb)
				time.Sleep(time.Second)
			}
		}()
	}
}

// double is a through stream that doubles values.
func through(cbMap pullstream.SourceCallbackMap) pullstream.Through {
	return func(read pullstream.Read) pullstream.Read {
		return func(end pullstream.EndOrError, cb pullstream.SourceCallback) {
			read(end, cbMap(cb))
		}
	}
}

func main() {
	cbLog := func(end pullstream.EndOrError, data interface{}) {
		if end.Yes() {
			return
		}
		log.WithField("data", data).Info("read")
	}
	// a logging sink
	sink := sink(cbLog)

	// a pushable source
	source, end, push := pullpushable.Pushable("example", func(end pullstream.EndOrError) {
		log.WithField("end", end).Debug("closed")
	})

	// a double through
	cbDouble := func(cb pullstream.SourceCallback) pullstream.SourceCallback {
		return func(end pullstream.EndOrError, data interface{}) {
			if v, ok := data.(int); ok {
				cb(pullstream.Null, v*2)
			}
		}
	}
	double := through(cbDouble)

	push(1)
	push(2)
	pullstream.Pull(source, double, sink)
	push(3)
	push(4)
	end(pullstream.Null)

	// queue source
	queue := queue.New()
	queue.Add("a")
	queue.Add("B")
	queue.Add("c")

	// an upper through
	cbToUpper := func(cb pullstream.SourceCallback) pullstream.SourceCallback {
		return func(end pullstream.EndOrError, data interface{}) {
			if v, ok := data.(string); ok {
				cb(pullstream.Null, strings.ToUpper(v))
			}
		}
	}
	upper := through(cbToUpper)

	pullstream.Pull(sourceQueue(queue), upper, sink)
	time.Sleep(time.Hour)
}
