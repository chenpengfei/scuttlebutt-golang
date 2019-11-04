package pkg

import (
	"context"
	"time"
)

type Stream struct {
	ctx context.Context
	Scuttlebutt
	// 连接到我的所有流节点
	peerStreams map[string]*Stream
}

func NewStream(ctx context.Context, sb Scuttlebutt) *Stream {
	return &Stream{
		ctx:         ctx,
		Scuttlebutt: sb,
		peerStreams: make(map[string]*Stream),
	}
}

// 连接对方流节点
func (s *Stream) Link(peerStream *Stream) {
	s.peerStreams[peerStream.ID] = peerStream
	s.startListening()
}

// 通知所有连接的对方节点，关闭己方流节点
func (s *Stream) Close() {
	s.broadcast(SbClose, nil)
	s.peerStreams = make(map[string]*Stream)
}

// 对方流节点消息更新，更新己方状态/数据
func (s *Stream) onEvent(nodeId string, sbEvent int, payload interface{}) {
	switch sbEvent {
	case SbClock:
		if v, ok := payload.(time.Time); ok {
			s.unicast(nodeId, SbUpdate, s.History(v))
		}
	case SbUpdate:
		if payload != nil {
			if v, ok := payload.(*Update); ok {
				s.ApplyUpdates(v)
			}
		}
	case SbClose:
		delete(s.peerStreams, nodeId)
	}
}

// 广播给对方流节点
func (s *Stream) broadcast(event int, payload interface{}) {
	for _, v := range s.peerStreams {
		v.onEvent(s.ID, event, payload)
	}
}

// 单播给对方节点
func (s *Stream) unicast(nodeId string, event int, payload interface{}) {
	if v, ok := s.peerStreams[nodeId]; ok {
		v.onEvent(nodeId, event, payload)
	}
}

func (s *Stream) startListening() {
	t := time.NewTicker(3 * time.Second)
	go func(refreshTicker *time.Ticker) {
		defer refreshTicker.Stop()
		for {
			select {
			case <-refreshTicker.C:
				s.broadcast(SbClock, s.Timestamp)
			case <-s.ctx.Done():
				return
			}
		}
	}(t)
}
