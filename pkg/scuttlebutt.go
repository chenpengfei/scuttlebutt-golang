package pkg

import "time"

const (
	SbClock = iota
	SbUpdate
	SbClose
)

// 流节点
type Node struct {
	// 全局唯一的节点ID
	ID string `json:"id"`
	// "我"所知道的所有和“我”相连的节点(包括我自己的)的最新时间戳
	// 会随着后续所有 stream 上收到的 update 来更新
	Timestamp time.Time `json:"timestamp"`
}

// 三元组
type Ternary struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// 一次更新
type Update struct {
	Ternarys []*Ternary
}

// Scuttlebutt 协议接口
type Protocol interface {
	// 更新己方消息
	ApplyUpdates(*Update)
	// 根据对端传来的 clock，计算出来的 delta。而 delta 是 update 是集合
	// peerClock: 每个 stream 上记录对端传来的 clock，并且会随着后续从 stream 收到的 update 不断更新
	History(peerClock time.Time) *Update
}

type Scuttlebutt struct {
	Protocol
	Node
}
