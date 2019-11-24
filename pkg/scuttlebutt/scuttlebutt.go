package scuttlebutt

import (
	event "github.com/chenpengfei/events/pkg/emitter"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	"github.com/sirupsen/logrus"
	"time"
)

// 全局唯一
type SourceId string

// "我"所知道的所有和“我”相连的节点(包括我自己的)的最新时间戳
// 会随着后续所有 stream 上收到的 Update 来更新
type Timestamp int64

type Sources map[SourceId]Timestamp

// 一次更新
type Update struct {
	SourceId  SourceId    `json:"source_id"`
	Timestamp Timestamp   `json:"timestamp"`
	From      SourceId    `json:"from"`
	Digest    string      `json:"digest"`
	Data      interface{} `json:"data"`
}

type Sign func(update *Update) (string, error)
type Verify func(update *Update) bool
type IsAccepted func(update *Update) bool

type ModelAccept interface {
	Whitelist() []string
	Blacklist() []string
}

type Protocol interface {
	IsAccepted(peerAccept interface{}, update *Update) bool
	// 更新己方消息
	ApplyUpdates(update *Update) bool
	// 根据对端传来的 clock，计算出来的 delta。而 delta 是 Update 集合
	// 每个 stream 上记录对端传来的 clock，并且会随着后续从 stream 收到的 Update 不断更新
	History(peerSources Sources, accept interface{}) []*Update
}

type Scuttlebutt struct {
	Protocol
	*event.Emitter

	id      SourceId
	Accept  interface{}
	Sources Sources

	sign     Sign
	verify   Verify
	createId CreateIdFn
	Streams  int

	log *logrus.Entry
}

func NewScuttlebutt(protocol Protocol, opts ...Option) *Scuttlebutt {
	sb := &Scuttlebutt{
		Protocol: protocol,
		Sources:  make(map[SourceId]Timestamp),
		Emitter:  event.NewEmitter(),
	}

	for _, opt := range opts {
		opt(sb)
	}

	if sb.id == "" {
		if sb.sign != nil && sb.verify != nil && sb.createId != nil {
			sb.id = sb.createId()
		} else {
			sb.id = CreateId()
		}
		if sb.id == "" {
			panic("id needed!")
		}
	}

	sb.log = logger.WithNamespace(string(sb.id))

	return sb
}

func (sb *Scuttlebutt) IsAccepted(peerAccept interface{}, update *Update) bool {
	panic("method(IsAccepted) must be implemented")
}

func (sb *Scuttlebutt) ApplyUpdates(update *Update) bool {
	panic("method(applyUpdate) must be implemented")
}

func (sb *Scuttlebutt) History(peerSources Sources, accept interface{}) []*Update {
	panic("method(history) must be implemented")
}

func (sb *Scuttlebutt) Id() SourceId {
	return sb.id
}

func (sb *Scuttlebutt) Update(update *Update) bool {
	sb.log.WithField("data", update.Data).Info("_update")

	ts := update.Timestamp
	sourceId := update.SourceId
	latest := sb.Sources[sourceId]

	if latest >= ts {
		sb.log.WithField("latest", latest).WithField("ts", ts).WithField("diff", latest-ts).Debug("Update is older, ignore it")
		sb.Emit("old_data", update)
		return false
	}

	sb.Sources[sourceId] = ts
	sb.log.WithField("sources", sb.Sources).Debug("update our sources to")

	if sourceId != sb.id {
		if sb.verify != nil {
			return sb.didVerification(sb.verify(update), update)
		} else {
			return sb.didVerification(true, update)
		}
	} else {
		if sb.sign != nil {
			if digest, err := sb.sign(update); err == nil {
				update.Digest = digest
			} else {
				return false
			}
		}
		return sb.didVerification(true, update)
	}
}

func (sb *Scuttlebutt) didVerification(verified bool, update *Update) bool {
	// I'm not sure how what should happen if an async verification
	// errors. if it's a key not found — that is a verification fail,
	// not an error. if its genuine error, really you should queue and.
	// try again? or replay the message later.
	// -- this should be done my the security plugin though, not scuttlebutt.
	if !verified {
		sb.Emit("unverified_data", update)
		return false
	}

	// emit '_update' event to notify every streams on this SB
	//todo.异步 ？
	r := sb.Protocol.ApplyUpdates(update)
	if r {
		sb.Emit("_update", update)
		sb.log.WithField("total_listeners", sb.ListenerCount("_update")).Debug("applied 'update' and fired ⚡_update")
	}
	return r
}

func (sb *Scuttlebutt) LocalUpdate(any interface{}) {
	sb.Update(&Update{
		SourceId:  sb.id,
		Timestamp: Timestamp(time.Now().UnixNano() / int64(time.Millisecond)),
		Data:      any,
	})
}

//each stream will be ended due to this event
func (sb *Scuttlebutt) Dispose() {
	sb.Emit("dispose", pull.End)
}
