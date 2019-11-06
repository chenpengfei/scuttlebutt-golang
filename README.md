# Scuttlebutt Golang 
尝试解决两个一致性问题
1. 最终一致性问题中的 “最终结果同步”，用一个分布式 "HashMap" 解决
2. “过程完全回放” 问题，用一个分布式 "Queue" 解决

## Usage

```
ctx := context.Background()

cs := sb.NewStream(ctx, sb.Scuttlebutt{
    Protocol: sb.NewModel("Client"),
    Node: sb.Node{
        ID:        "XXXX",
        Timestamp: time.Now(),
    },
})

ss := sb.NewStream(ctx, sb.Scuttlebutt{
    Protocol: sb.NewModel("Server"),
    Node: sb.Node{
        ID:        "YYYY",
        Timestamp: time.Now(),
    },
})

cs.Link(ss)
ss.Link(cs)
```

## Run
```
make run
```




