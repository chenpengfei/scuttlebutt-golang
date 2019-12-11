package duplex

import (
	"errors"
	event "github.com/chenpengfei/events/pkg/emitter"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	sb "github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"github.com/sirupsen/logrus"
)

// 每一个 SB 维护自己从所有 streams 获取到的知识 (clock)和 updates
// 由所有 history 和后续 updates 累计起来

// 节点一旦收到 update (无论是自己发起的还是传入的)，都将广播给其他链接到该节点的 streams

// 每一个 stream 上，记录和维护了对端的 peerClock (对端的知识)，
// 除了初始的SYNC，stream 上收到的全部 peer 发来(或者转发)的 update，
// 也将触发 peerClock 的更新，这样 SB 才能知道其他节点来的 update 是否需要通过这个 stream 转到 peer

// stream 的读写是可控的，因此，SB 节点之间的网状拓扑的形式可以很丰富

// Stream 是双工协议的抽象

type OnClose func(err error)

func validate(update *sb.Update) bool {
	if update == nil {
		return false
	} else {
		return true
	}
}

type Outgoing struct {
	Id     sb.SourceId
	Clock  sb.Sources
	Meta   interface{}
	Accept interface{}
}

type Stream interface {
	Source() interface{}
	Sink() interface{}
}

type Duplex struct {
	sb *sb.Scuttlebutt
	*event.Emitter

	name        string
	source      interface{}
	sink        interface{}
	wrapper     string
	readable    bool
	writable    bool
	ended       error
	abort       error
	syncSent    bool
	syncRecv    bool
	buffer      *Buffer
	cb          pull.SourceCallback
	onclose     OnClose
	isFirstRead bool
	tail        bool
	peerSources sb.Sources
	peerAccept  interface{}
	peerId      sb.SourceId
	meta        interface{}
	log         *logrus.Entry
}

// Sink (reader or writable stream that consumes values)
// Source (readable stream that produces values)
func NewDuplex(sb *sb.Scuttlebutt, opts ...Option) *Duplex {
	duplex := &Duplex{
		sb:          sb,
		Emitter:     event.NewEmitter(),
		name:        "stream",
		wrapper:     "json",
		readable:    true,
		writable:    true,
		ended:       nil,
		abort:       nil,
		syncSent:    false,
		syncRecv:    false,
		buffer:      NewBuffer(),
		cb:          nil,
		onclose:     nil,
		isFirstRead: true,
		tail:        true,
		peerSources: nil,
		peerAccept:  nil,
		peerId:      "",
	}

	for _, opt := range opts {
		opt(duplex)
	}

	duplex.log = logger.WithNamespace(duplex.name)

	// Non-writable means we could skip receiving SYNC from peer
	duplex.syncRecv = !duplex.writable

	// Non-readable means we don't need to send SYNC to peer
	duplex.syncSent = !duplex.readable

	duplex.sb = sb

	duplex.onclose = func(err error) {
		duplex.sb.RemoveListener("_update", duplex.onUpdate)
		duplex.sb.RemoveListener("dispose", duplex.End)
		duplex.sb.Streams--
		duplex.Emit("unstream", duplex.sb.Streams)
	}

	sb.Streams++
	sb.Once("dispose", duplex.End)

	return duplex
}

func (d *Duplex) drain() {
	if d.cb == nil {
		// there is no downstream waiting for callback
		if d.ended != nil && d.onclose != nil {
			// perform _onclose regardless of whether there is data in the cache
			c := d.onclose
			d.onclose = nil
			c(d.ended)
		}
		return
	}

	if d.abort != nil {
		// downstream is waiting for abort
		d.callback(d.abort, nil)
	} else if d.buffer.Length() == 0 && d.ended != nil {
		// we'd like to end and there is no left items to be sent
		d.callback(d.ended, nil)
	} else if d.buffer.Length() != 0 {
		d.callback(nil, d.buffer.Shift())
	}
}

func (d *Duplex) callback(err error, data interface{}) {
	_cb := d.cb
	if err != nil && d.onclose != nil {
		c := d.onclose
		d.onclose = nil
		c(err)
	}
	d.cb = nil
	if _cb != nil {
		_cb(err, data)
	}
}

func (d *Duplex) getOutgoing() *Outgoing {
	outgoing := &Outgoing{
		Id:    d.sb.Id(),
		Clock: d.sb.Sources,
	}

	if d.sb.Accept != nil {
		outgoing.Accept = d.sb.Accept
	}

	if d.meta != nil {
		outgoing.Meta = d.meta
	}

	return outgoing
}

// process any update ocurred on sb
func (d *Duplex) onUpdate(data interface{}) {
	update := data.(*sb.Update)
	d.log.WithField("update", update).Debug("got 'update' on stream:")

	// current stream is in write-only mode
	if !d.readable {
		d.log.Debug("'update' ignored by its non-readable flag")
		return
	}

	if !validate(update) || !sb.Filter(update, d.peerSources) {
		return
	}

	// this update comes from our peer stream, don't send back
	if update.From == d.peerId {
		d.log.WithField("peerId", d.peerId).Debug("'update' ignored by peerId:")
		return
	}

	isAccepted := true
	if d.peerAccept != nil {
		isAccepted = d.sb.Protocol.IsAccepted(d.peerAccept, update)
	}

	if !isAccepted {
		d.log.WithField("update", update).WithField("peerAccept", d.peerAccept).Debug("'update' ignored by peerAccept")
		return
	}

	// send 'scuttlebutt' to peer
	update.From = d.sb.Id()
	d.push(update, false)
	d.log.WithField("update", update).Debug("'sent 'update to peer")

	// really, this should happen before emitting.
	ts := update.Timestamp
	source := update.SourceId
	d.peerSources[source] = ts
	d.log.WithField("peerSources", d.peerSources).Debug("updated peerSources to")
}

func (d *Duplex) rawSource(abort error, cb pull.SourceCallback) {
	if abort != nil {
		d.abort = abort
		// if there is already a cb waiting, abort it.
		if d.cb != nil {
			d.callback(abort, nil)
		}
	}

	if d.isFirstRead {
		d.isFirstRead = false
		outgoing := d.getOutgoing()
		d.push(outgoing, true)
		d.log.WithField("outgoing", outgoing).Debug("sent 'outgoing'")
	}

	d.cb = cb
	d.drain()
}

func (d *Duplex) rawSink(read pull.Read) {
	var next pull.SourceCallback
	next = func(end error, update interface{}) {
		if end == pull.ErrPullStreamEnd {
			d.log.Debugf("sink ended by peer(%v)", d.peerId)
			d.End(end)
			return
		}

		if end != nil {
			d.log.Debugf("sink reading errors: %v", end)
			d.End(end)
			return
		}

		if v, ok := update.(*sb.Update); ok {
			d.log.WithField("update", v.Data).Debugf("sink reads data from peer(%v)", d.peerId)
			if !d.writable {
				return
			}
			if validate(v) {
				d.sb.Update(v)
			}
		} else if v, ok := update.(string); ok {
			d.log.WithField("update", v).Debugf("sink reads data from peer(%v)", d.peerId)
			cmd := v
			if d.writable {
				if "SYNC" == cmd {
					d.log.Info("SYNC received")
					d.syncRecv = true
					d.Emit("syncReceived", nil)
					if d.syncSent {
						d.log.Info("emit synced")
						d.Emit("synced", nil)
					}
				}
			} else {
				d.log.Infof("ignore peer's(%v) SYNC due to our non-writable setting", d.peerId)
			}
		} else if v, ok := update.(*Outgoing); ok {
			// it's a scuttlebutt digest(vector clocks) when Clock is an object.
			if d.readable {
				d.log.WithField("update", v).Debugf("sink reads data from peer(%v)", v.Id)
				d.start(v)
			} else {
				d.peerId = v.Id
				d.log.Infof("ignore peer's(%v) outgoing data due to our non-readable setting", v.Id)
			}
		} else {
			d.Emit("error", nil)
			d.End(errors.New("unknown data type"))
		}
		read(d.endOrError(), next)
	}
	read(d.endOrError(), next)
	//todo.通过 channel 实现，防止栈溢出
}

func (d *Duplex) endOrError() error {
	if d.ended != nil {
		return d.ended
	}
	return d.abort
}

func (d *Duplex) Source() interface{} {
	if d.source == nil {
		if d.wrapper == "raw" {
			d.source = pull.Read(d.rawSource)
		} else if d.wrapper == "json" {
			d.source = pull.Pull(pull.Read(d.rawSource), Serialize())
		} else {
			d.source = pull.Read(d.rawSource)
		}
	}
	return d.source
}

func (d *Duplex) Sink() interface{} {
	if d.sink == nil {
		if d.wrapper == "raw" {
			d.sink = pull.Sink(d.rawSink)
		} else if d.wrapper == "json" {
			d.sink = pull.Pull(Parse(), pull.Sink(d.rawSink))
		} else {
			d.sink = pull.Sink(d.rawSink)
		}
	}
	return d.sink
}

//todo.可能会爆
func (d *Duplex) push(data interface{}, toHead bool) {
	if d.ended != nil {
		return
	}

	// if sink already waiting,
	// we can call back directly.
	if d.cb != nil {
		d.callback(d.abort, data)
		return
	}

	// otherwise, buffer data
	if toHead {
		d.buffer.Unshift(data)
	} else {
		d.buffer.Push(data)
	}
}

func (d *Duplex) End(data interface{}) {
	end := data.(error)
	if d.ended == nil {
		if end != nil {
			d.ended = end
		} else {
			d.ended = pull.ErrPullStreamEnd
		}
	}

	// attempt to drain
	d.drain()
}

func (d *Duplex) start(incoming *Outgoing) {
	d.log.WithField("incoming", incoming).Info("start with incoming")
	if incoming.Clock == nil {
		d.Emit("error", nil)
		d.End(pull.ErrPullStreamEnd)
		return
	}

	d.peerSources = incoming.Clock
	d.peerId = incoming.Id
	d.peerAccept = incoming.Accept

	rest := func() {
		// when we have sent all history
		d.Emit("header", incoming)
		// when we have received all history
		// emit 'synced' when this stream has synced.
		if d.syncRecv {
			d.log.Info("emit synced")
			d.Emit("synced", nil)
		}
		if !d.tail {
			d.End(pull.ErrPullStreamEnd)
		}
	}

	// won't send history/SYNC and further update out if the stream is write-only
	if d.readable {
		// call this.history to calculate the delta between peers
		// AsyncScuttlebutt
		history := d.sb.Protocol.History(d.peerSources, d.peerAccept)
		for _, h := range history {
			h.From = d.sb.Id()
			d.push(h, false)
			d.log.WithField("history", h).Debugf("'history' to peer(%v) has been sent", d.peerId)
		}
		d.sb.On("_update", d.onUpdate)

		d.push("SYNC", false)
		d.syncSent = true
		d.log.Debugf("sent 'SYNC' to peer(%v)", d.peerId)
	}

	rest()
}
