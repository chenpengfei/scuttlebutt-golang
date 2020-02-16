package model

import (
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	log "github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	sb "github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
)

type ValueModel struct {
	K string
	V interface{}
}

type ValueModelFrom struct {
	*ValueModel
	From sb.SourceId
}

// MemoryModel 的操作不考虑“同步”
// 同步由 MemoryModel 间 Stream 负责
// 即业务和 IO 分离
type SyncModel struct {
	*sb.Scuttlebutt
	store map[string]*sb.Update
}

func NewSyncModel(opts ...sb.Option) *SyncModel {
	model := &SyncModel{store: make(map[string]*sb.Update)}
	model.Scuttlebutt = sb.NewScuttlebutt(model, opts...)
	return model
}

func (s *SyncModel) IsAccepted(peerAccept *sb.ModelAccept, update *sb.Update) bool {
	if peerAccept != nil {
		k := update.Data[sb.ModelValueItemsKey]
		if peerAccept.Blacklist != nil {
			for _, b := range peerAccept.Blacklist {
				if k == b {
					return false
				}
			}
		}
		if peerAccept.Whitelist != nil {
			for _, w := range peerAccept.Whitelist {
				if k == w {
					return true
				}
			}
			return false
		}
	}
	return true
}

func (s *SyncModel) ApplyUpdates(update *sb.Update) bool {
	k := update.Data[sb.ModelValueItemsKey].(string)
	v := update.Data[sb.ModelValueItemsValue]

	// ignore if we already have a more recent value
	if v, found := s.store[k]; found && v.Timestamp > update.Timestamp {
		s.Emit("_remove", update)
		return false
	}

	if s.store[k] != nil {
		s.Emit("_remove", s.store[k])
	}

	s.store[k] = update
	s.Emit("update", update)
	s.Emit("changed", &ValueModel{K: k, V: v})
	s.Emit("changed:"+k, v)

	if s.Id() != update.SourceId {
		s.Emit("changedByPeer", &ValueModelFrom{
			ValueModel: &ValueModel{K: k, V: v},
			From:       update.From,
		})
	}

	return true
}

func (s *SyncModel) History(peerSources sb.Sources, accept *sb.ModelAccept) []*sb.Update {
	h := make([]*sb.Update, 0)
	for _, update := range s.store {
		if accept != nil && !s.IsAccepted(accept, update) {
			continue
		}
		if sb.Filter(update, peerSources) {
			h = append(h, update)
		}
	}
	sb.Sort(h)
	return h
}

func (s *SyncModel) Set(k string, v interface{}) *SyncModel {
	log.WithField("k", k).WithField("v", v).Debug("set")

	s.LocalUpdate([2]interface{}{k, v})
	return s
}

func (s *SyncModel) Get(k string, withClock bool) interface{} {
	if _, found := s.store[k]; found {
		if withClock {
			return s.store[k]
		} else {
			return s.store[k].Data[sb.ModelValueItemsValue]
		}
	}
	return nil
}

func (s *SyncModel) Keys() []string {
	slice := make([]string, 0)
	for k := range s.store {
		if s.Get(k, false) != nil {
			slice = append(slice, k)
		}
	}
	return slice
}

func (s *SyncModel) CreateStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(s.Scuttlebutt, opts...)
}

func (s *SyncModel) CreateWriteStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(
		s.Scuttlebutt,
		duplex.WithWritable(true),
		duplex.WithReadable(false))
}

func (s *SyncModel) CreateSinkStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(
		s.Scuttlebutt,
		duplex.WithWritable(true),
		duplex.WithReadable(false))
}

func (s *SyncModel) CreateReadStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(
		s.Scuttlebutt,
		duplex.WithWritable(false),
		duplex.WithReadable(true))
}

func (s *SyncModel) CreateSourceStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(
		s.Scuttlebutt,
		duplex.WithWritable(false),
		duplex.WithReadable(true))
}

var _ sb.Model = new(SyncModel)
