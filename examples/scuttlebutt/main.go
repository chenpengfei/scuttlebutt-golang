package main

import (
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	log "github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/model"
	sb "github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"time"
)

func main() {
	a := model.NewSyncModel(sb.WithId("A"))
	b := model.NewSyncModel(sb.WithId("B"))

	sa := a.CreateStream(duplex.WithName("a->b"))
	sb := b.CreateStream(duplex.WithName("b->a"))

	a.Set("foo", "changed by A")

	sb.On("synced", func(data interface{}) {
		printKeyValue(b, "foo")
	})

	duplex.Link(sa, sb)

	time.Sleep(100 * time.Millisecond)
}

func printKeyValue(model *model.SyncModel, key string) {
	log.WithField("id", model.Id()).WithField("value", model.Get(key, false)).Info("with clock")
	log.WithField("id", model.Id()).WithField("value", model.Get(key, true)).Info("with no clock")
}
