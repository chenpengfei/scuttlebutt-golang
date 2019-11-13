package main

import (
	"scuttlebutt-golang/pkg/duplex"
	"scuttlebutt-golang/pkg/model"
	sb "scuttlebutt-golang/pkg/scuttlebutt"
	"time"
)

func main() {
	a := model.NewSyncModel(sb.WithId("A"))
	b := model.NewSyncModel(sb.WithId("B"))

	sa := a.CreateStream(duplex.WithName("a->b"))
	sb := b.CreateStream(duplex.WithName("b->a"))

	a.Set("foo", "changed by A")

	sb.On("synced", func(data interface{}) {
		PrintKeyValue(b, "foo")
	})

	duplex.Link(sa, sb)

	for {
		time.Sleep(time.Second)
	}
}
