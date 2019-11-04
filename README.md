# scuttlebutt-golang 
Scuttlebutt Golang 内存版本

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




