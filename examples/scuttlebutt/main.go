package main

import (
	"scuttlebutt-golang/pkg/duplex"
	log "scuttlebutt-golang/pkg/logger"
	"scuttlebutt-golang/pkg/model"
	sb "scuttlebutt-golang/pkg/scuttlebutt"
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
}

func printKeyValue(model *model.SyncModel, key string) {
	log.WithField("id", model.Id()).WithField("value", model.Get(key, false)).Info("with clock")
	log.WithField("id", model.Id()).WithField("value", model.Get(key, true)).Info("with no clock")
}
