package model

import (
	"scuttlebutt-golang/pkg/duplex"
	log "scuttlebutt-golang/pkg/logger"
	sb "scuttlebutt-golang/pkg/scuttlebutt"
)

type ValueModel struct {
	K string
	V interface{}
}

type ValueModelFrom struct {
	Value *ValueModel
	From  sb.SourceId
}

// MemoryModel 的操作不考虑“同步”
// 同步由 MemoryModel 间 Stream 负责
// 即业务和 IO 分离
type SyncModel struct {
	*sb.Scuttlebutt

	store map[string]*sb.Update
}

func (s *SyncModel) Id() string {
	return string(s.Scuttlebutt.Id)
}

func (s *SyncModel) ApplyUpdates(update *sb.Update) bool {
	key := update.Data.(*ValueModel).K

	// ignore if we already have a more recent value
	if v, found := s.store[key]; found && v.Timestamp > update.Timestamp {
		s.Emit("_remove", update)
		return false
	}

	if s.store[key] != nil {
		s.Emit("_remove", s.store[key])
	}

	s.store[key] = update
	//qa. this.emit.apply(this, ['update', update])
	s.Emit("update", update)
	s.Emit("changed:"+key, update.Data.(*ValueModel).V)

	if s.Id() != string(update.SourceId) {
		s.Emit("changedByPeer:"+key+", from:"+string(update.From), update.Data.(*ValueModel).V)
	}

	return true
}

func (s *SyncModel) History(peerSources sb.Sources) []*sb.Update {
	h := make([]*sb.Update, 0)
	for _, update := range s.store {
		if sb.Filter(update, peerSources) {
			h = append(h, update)
		}
	}
	sb.Sort(h)
	return h
}

func (s *SyncModel) Set(k string, v interface{}) *SyncModel {
	log.WithField("k", k).WithField("v", v).Debug("set")
	s.LocalUpdate(&ValueModel{K: k, V: v})
	return s
}

func (s *SyncModel) Get(k string, withClock bool) interface{} {
	if _, found := s.store[k]; found {
		if withClock {
			return s.store[k]
		} else {
			return s.store[k].Data.(*ValueModel).V
		}
	}
	return nil
}

func (s *SyncModel) Keys() []string {
	slice := make([]string, 0)
	for k, _ := range s.store {
		if s.Get(k, false) != nil {
			slice = append(slice, k)
		}
	}
	return slice
}

func (s *SyncModel) CreateStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(s.Scuttlebutt, opts...)
}

func (s *SyncModel) CreateSinkStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(s.Scuttlebutt, duplex.WithWritable(true), duplex.WithReadable(false))
}

func (s *SyncModel) CreateSourceStream(opts ...duplex.Option) *duplex.Duplex {
	return duplex.NewDuplex(s.Scuttlebutt, duplex.WithWritable(false), duplex.WithReadable(true))
}

func NewSyncModel(opts ...sb.Option) *SyncModel {
	model := &SyncModel{store: make(map[string]*sb.Update)}
	model.Scuttlebutt = sb.NewScuttlebutt(model, opts...)
	return model
}

var _ sb.Protocol = new(SyncModel)
