package duplex

import (
	"github.com/eapache/queue"
	"github.com/sirupsen/logrus"
	"scuttlebutt-golang/pkg/event"
	"scuttlebutt-golang/pkg/logger"
	pullstream "scuttlebutt-golang/pkg/pull-stream"
	sb "scuttlebutt-golang/pkg/scuttlebutt"
)

// 每一个 SB 维护自己从所有 streams 获取到的知识 (clock)和 updates
// 由所有 history 和后续 updates 累计起来

// 节点一旦收到 update (无论是自己发起的还是传入的)，都将广播给其他链接到该节点的 streams

// 每一个 stream 上，记录和维护了对端的 peerClock (对端的知识)，
// 除了初始的SYNC，stream 上收到的全部 peer 发来(或者转发)的 update，
// 也将触发 peerClock 的更新，这样 SB 才能知道其他节点来的 update 是否需要通过这个 stream 转到 peer

// stream 的读写是可控的，因此，SB 节点之间的网状拓扑的形式可以很丰富

// Stream 是双工协议的抽象

type OnClose func(err pullstream.EndOrError)

func validate(update *sb.Update) bool {
	if update == nil {
		return false
	} else {
		return true
	}
}

type Outgoing struct {
	id     sb.SourceId
	clock  sb.Sources
	meta   interface{}
	accept interface{}
}

type Duplex struct {
	sb *sb.Scuttlebutt
	*event.Event

	name        string
	source      pullstream.Read
	sink        pullstream.Sink
	wrapper     string
	readable    bool
	writable    bool
	ended       pullstream.EndOrError
	abort       pullstream.EndOrError
	syncSent    bool
	syncRecv    bool
	buffer      *queue.Queue
	cb          pullstream.SourceCallback
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
		Event:       event.NewEvent(),
		name:        "stream",
		wrapper:     "json",
		readable:    true,
		writable:    true,
		ended:       pullstream.Null,
		abort:       pullstream.Null,
		syncSent:    false,
		syncRecv:    false,
		buffer:      queue.New(),
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

	sb.Streams++
	sb.On("dispose", duplex.end)

	return duplex
}

func (d *Duplex) onClose() {
	d.sb.RemoveListener("_update", d.onUpdate)
	d.sb.RemoveListener("dispose", d.end)
	d.sb.Streams--
	d.Emit("unstream", d.sb.Streams)
}

func (d *Duplex) drain() {
	if d.cb == nil {
		// there is no downstream waiting for callback
		if d.ended.Yes() && d.onclose != nil {
			// perform _onclose regardless of whether there is data in the cache
			c := d.onclose
			c = nil
			c(d.ended)
		}
		return
	}

	if d.abort.Yes() {
		// downstream is waiting for abort
		d.callback(d.abort, nil)
	} else if d.buffer.Length() == 0 && d.ended.Yes() {
		// we'd like to end and there is no left items to be sent
		d.callback(d.ended, nil)
	} else if d.buffer.Length() != 0 {
		d.callback(pullstream.Null, d.buffer.Remove())
	}
}

func (d *Duplex) callback(err pullstream.EndOrError, data interface{}) {
	cb := d.cb
	if err.Yes() && d.onclose != nil {
		c := d.onclose
		d.onclose = nil
		//qa. c(err === true ? null : err)
		c(err)
	}
	d.cb = nil
	if cb != nil {
		cb(err, data)
	}
}

func (d *Duplex) getOutgoing() *Outgoing {
	outgoing := &Outgoing{
		id:    d.sb.Id,
		clock: d.sb.Sources,
	}

	if d.sb.Accept != nil {
		outgoing.accept = d.sb.Accept
	}

	if d.meta != nil {
		outgoing.meta = d.meta
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
		isAccepted = d.sb.IsAccepted(d.peerAccept, update)
	}

	if !isAccepted {
		d.log.WithField("update", update).WithField("peerAccept", d.peerAccept).Debug("'update' ignored by peerAccept")
		return
	}

	// send 'scuttlebutt' to peer
	update.From = d.sb.Id
	d.push(update, false)
	d.log.WithField("update", update).Debug("'sent 'update to peer")

	// really, this should happen before emitting.
	ts := update.Timestamp
	source := update.SourceId
	d.peerSources[source] = ts
	d.log.WithField("peerSources", d.peerSources).Debug("updated peerSources to")
}

func (d *Duplex) rawSource(abort pullstream.EndOrError, cb pullstream.SourceCallback) {
	//qa. if (abort) { 代表 End 或 Error.
	if abort.Yes() {
		d.abort = abort
		// if there is already a cb waiting, abort it.
		if cb != nil {
			cb(abort, nil)
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

func (d *Duplex) rawSink(read pullstream.Read) {
	var next pullstream.SourceCallback
	next = func(end pullstream.EndOrError, update interface{}) {
		if end.End() {
			d.log.WithField("peerId", d.peerId).Debug("sink ended by peer")
			d.end(end)
			return
		}

		if end.Error() {
			d.log.Debug("sink reading errors")
			d.end(end)
			return
		}

		d.log.WithField("update", update).WithField("peerId", d.peerId).Debug("sink reads data from peer")
		if v, ok := update.(*sb.Update); ok {
			if !d.writable {
				return
			}
			if validate(v) {
				d.sb.Update(v)
			}
		} else if v, ok := update.(string); ok {
			cmd := v
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
			// it's a scuttlebutt digest(vector clocks) when clock is an object.
			//qa. self.start(update).then(() => {
			d.start(update)
		}
		read(d.endOrError(), next)
	}
	read(d.endOrError(), next)
}

func (d *Duplex) endOrError() pullstream.EndOrError {
	if d.ended.Yes() {
		return d.ended
	}
	return d.abort
}

func (d *Duplex) GetSource() pullstream.Read {
	if d.source == nil {
		if d.wrapper == "raw" {
			d.source = d.rawSource
		} else {
			//todo
			d.source = d.rawSource
		}
	}
	return d.source
}

func (d *Duplex) GetSink() pullstream.Sink {
	if d.sink == nil {
		d.sink = d.rawSink
	} else {
		d.sink = d.rawSink
	}
	return d.sink
}

func (d *Duplex) push(data interface{}, toHead bool) {
	if d.ended.Yes() {
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
		d.buffer.Add(data)
	} else {
		d.buffer.Add(data)
	}
}

func (d *Duplex) end(data interface{}) {
	end := data.(pullstream.EndOrError)
	if !d.ended.Yes() {
		if end.Yes() {
			d.ended = end
		} else {
			d.ended = pullstream.End
		}
	}

	// attempt to drain
	d.drain()
}

func (d *Duplex) start(data interface{}) {
	d.log.WithField("data", data).Info("start")
	if incoming, ok := data.(*Outgoing); ok {
		d.peerSources = incoming.clock
		d.peerId = incoming.id
		d.peerAccept = incoming.accept

		rest := func() {
			d.push("SYNC", false)
			d.syncSent = true
			d.log.WithField("peerId", d.peerId).Debug("sent 'SYNC' to peer")

			// when we have sent all history
			d.Emit("header", incoming)
			d.Emit("syncSent", nil)
			// when we have received all history
			// emit 'synced' when this stream has synced.
			if d.syncRecv {
				d.Emit("synced", nil)
			}
			if !d.tail {
				d.end(pullstream.Null)
			}
		}

		// won't send history out if the stream is write-only
		if !d.readable {
			d.sb.On("_update", d.onUpdate)
			rest()
			return
		}

		// call this.history to calculate the delta between peers
		// AsyncScuttlebutt
		history := d.sb.Protocol.History(d.peerSources)
		for _, h := range history {
			h.From = d.sb.Id
			d.push(h, false)
			d.log.WithField("peerId", d.peerId).WithField("history", h).Debug("sent history")
		}
		d.sb.On("_update", d.onUpdate)
		//qa. sent history 此时应该是等对方发送 'SYNC'.
		rest()
	} else {
		d.Emit("error", nil)
		d.end(pullstream.Err)
		return
	}
}
