package model

import (
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStreamEnd(t *testing.T) {
	assert := assert.New(t)

	t.Run("end one stream", func(t *testing.T) {
		a := NewSyncModel(scuttlebutt.WithId("A"))
		b := NewSyncModel(scuttlebutt.WithId("B"))

		sa := a.CreateStream()
		sb := b.CreateStream()

		duplex.Link(sa, sb)

		assert.Equal(1, a.ListenerCount("_update"))
		assert.Equal(1, b.ListenerCount("_update"))

		sa.End(pull.Null)

		assert.Equal(0, a.ListenerCount("_update"))
		assert.Equal(0, b.ListenerCount("_update"))
	})

	t.Run("stream count", func(t *testing.T) {
		a := NewSyncModel()
		b := NewSyncModel()

		sa := a.CreateStream()
		sb := b.CreateStream()

		assert.Equal(1, a.Streams)
		assert.Equal(1, b.Streams)

		duplex.Link(sa, sb)

		a.On("unstream", func(data interface{}) {
			assert.Equal(0, data.(int))
		})
		b.On("unstream", func(data interface{}) {
			assert.Equal(0, data.(int))
		})

		sa.End(pull.Null)
	})

	t.Run("stream count", func(t *testing.T) {
		a := NewSyncModel()
		b := NewSyncModel()

		sa := a.CreateStream()
		sb := b.CreateStream()

		assert.Equal(1, a.Streams)
		assert.Equal(1, b.Streams)

		duplex.Link(sa, sb)

		a.On("dispose", func(data interface{}) {
			assert.Equal(pull.End, data)
		})

		a.Dispose()
	})
}
