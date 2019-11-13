package scuttlebutt

import (
	"github.com/sirupsen/logrus"
	"scuttlebutt-golang/pkg/event"
	"scuttlebutt-golang/pkg/logger"
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
	History(peerSources Sources) []*Update
}

type Scuttlebutt struct {
	Protocol
	*event.Event

	Id      SourceId
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
		Sources: make(map[SourceId]Timestamp),
		Event:   event.NewEvent(),
	}

	for _, opt := range opts {
		opt(sb)
	}

	if sb.sign != nil && sb.verify != nil {
		if sb.createId != nil {
			sb.Id = sb.createId()
		}
		if sb.Id == "" {
			panic("Id needed!")
		}
	} else {
		sb.Id = CreateId()
	}

	for _, opt := range opts {
		opt(sb)
	}

	sb.log = logger.WithNamespace(string(sb.Id))

	return sb
}

func (sb *Scuttlebutt) IsAccepted(peerAccept interface{}, update *Update) bool {
	panic("method(IsAccepted) must be implemented")
}

func (sb *Scuttlebutt) ApplyUpdates(update *Update) bool {
	panic("method(applyUpdate) must be implemented")
	return false
}

func (sb *Scuttlebutt) History(peerSources Sources) []*Update {
	panic("method(history) must be implemented")
	return []*Update{}
}

// localUpdate 和 history 会触发 Update
func (sb *Scuttlebutt) Update(update *Update) bool {
	sb.log.WithField("data", update.Data).Info("update")

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

	if sourceId != sb.Id {
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
		sb.log.Debug("applied 'update' and fired ⚡_update")
	}
	return r
}

func (sb *Scuttlebutt) LocalUpdate(any interface{}) {
	sb.Update(&Update{
		SourceId:  sb.Id,
		Timestamp: Timestamp(time.Now().UnixNano()),
		Data:      any,
	})
}

// each stream will be ended due to this event
func (sb *Scuttlebutt) dispose() {
	sb.Emit("dispose", nil)
}
