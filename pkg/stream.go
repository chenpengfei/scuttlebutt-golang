package pkg

import (
	"context"
	"time"
)

// 每一个 SB 维护自己从所有 streams 获取到的知识 (clock)和 updates
// 由所有 history 和后续 updates 累计起来

// 节点一旦收到 update (无论是自己发起的还是传入的)，都将广播给其他链接到该节点的 streams

// 每一个 stream 上，记录和维护了对端的 peerClock (对端的知识)，
// 除了初始的SYNC，stream 上收到的全部 peer 发来(或者转发)的 update，
// 也将触发 peerClock 的更新，这样 SB 才能知道其他节点来的 update 是否需要通过这个 stream 转到 peer

// stream 的读写是可控的，因此，SB 节点之间的网状拓扑的形式可以很丰富

// 双工协议的抽象
type Stream struct {
	ctx context.Context
	Scuttlebutt
	// 连接到我的流节点
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
		if p, ok := payload.(time.Time); ok {
			s.unicast(nodeId, SbUpdate, s.History(p))
		}
	case SbUpdate:
		if payload != nil {
			if p, ok := payload.(*Update); ok {
				s.ApplyUpdates(p)
				//s.broadcast(SbUpdate, payload)
			}
		}
	case SbClose:
		delete(s.peerStreams, nodeId)
	}
}

// 广播给对方流节点
func (s *Stream) broadcast(event int, payload interface{}) {
	for _, peer := range s.peerStreams {
		peer.onEvent(s.ID, event, payload)
	}
}

// 单播给对方节点
func (s *Stream) unicast(nodeId string, event int, payload interface{}) {
	if peer, ok := s.peerStreams[nodeId]; ok {
		peer.onEvent(nodeId, event, payload)
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
