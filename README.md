# Scuttlebutt Golang 
[![Build Status](https://travis-ci.org/chenpengfei/scuttlebutt-golang.svg)](https://travis-ci.org/chenpengfei/scuttlebutt-golang)
[![Coverage Status](https://coveralls.io/repos/github/chenpengfei/scuttlebutt-golang/badge.svg)](https://coveralls.io/github/chenpengfei/scuttlebutt-golang)


尝试解决两个一致性问题
1. 最终一致性问题中的 “最终结果同步”，用一个分布式 "HashMap" 解决
2. “过程完全回放” 问题，用一个分布式 "Queue" 解决

## Usage

```
a := model.NewSyncModel(sb.WithId("A"))
b := model.NewSyncModel(sb.WithId("B"))

sa := a.CreateStream(duplex.WithName("a->b"))
sb := b.CreateStream(duplex.WithName("b->a"))

a.Set("foo", "changed by A")

sb.On("synced", func(data interface{}) {
    PrintKeyValue(b, "foo")
})

duplex.Link(sa, sb)
```

## Run
```
make run
```

# Readable/Writable
## Read-only, !writeable 的逻辑
### 要求
1. 本端的状态不会被 Peer 改变
2. 会接收 Peer 的 Clock，且传递相应的 History 给 Peer
   1. 所以仅仅是 pull(a.source, b.sink) 将不能工作，因为读不到 Peer 的 Clock
   2. 还有一种粗暴的实现就是不接收 Peer Clock，直接返回全部 History，这样的缺点是躲过 Peer 的 SB 已经从其他节点同步了一部分
      内容，这个全量就有些浪费带宽了
3. 忽视 Peer 所有发来的 Update，也不更新 PeerClock
4. 忽视对方发来的 Sync，因为缺省就是 _syncRecv = true

## Write-only, !readable 的逻辑
### 要求
1. 本端会丢弃 Peer 的 Outgoing，因此
   1. 不会发送 History
   2. 不会发送 Sync，缺省设置 _syncSent = true
   3. 不会发出 Update，即不监听 SB 上的 _update
2. SYNC 还是会接收，这样可以正常产生 synced 事件





