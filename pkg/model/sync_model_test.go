package model

import (
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	sb "github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncModel(t *testing.T) {
	assert := assert.New(t)

	expected := struct {
		key    string
		valueA string
		valueB string
	}{
		key:    "foo",
		valueA: "changed by A",
		valueB: "changed by B",
	}

	t.Run("local change", func(t *testing.T) {
		a := NewSyncModel(sb.WithId("A"))
		a.On("changed", func(data interface{}) {
			assert.Equal(expected.key, data.(*ValueModel).K)
			assert.Equal(expected.valueA, data.(*ValueModel).V)
		})
		a.On("changed"+expected.key, func(data interface{}) {
			assert.Equal(expected.valueA, data.(string))
		})
		a.Set(expected.key, expected.valueA)
	})

	t.Run("change before sync", func(t *testing.T) {
		a := NewSyncModel(sb.WithId("A"))
		b := NewSyncModel(sb.WithId("B"))

		sa := a.CreateStream(duplex.WithName("a->b"))
		sb := b.CreateStream(duplex.WithName("b->a"))

		a.Set(expected.key, expected.valueA)

		sb.On("synced", func(data interface{}) {
			assert.Equal(expected.valueA, b.Get(expected.key, false).(string))
		})

		duplex.Link(sa, sb)
	})

	t.Run("change after sync", func(t *testing.T) {
		a := NewSyncModel(sb.WithId("A"))
		b := NewSyncModel(sb.WithId("B"))

		sa := a.CreateStream(duplex.WithName("a->b"))
		sb := b.CreateStream(duplex.WithName("b->a"))

		b.On("changedByPeer", func(data interface{}) {
			assert.Equal(a.Id(), data.(*ValueModelFrom).From)
			assert.Equal(expected.key, data.(*ValueModelFrom).K)
			assert.Equal(expected.valueA, data.(*ValueModelFrom).V)
			assert.Equal(expected.valueA, b.Get(expected.key, false))
		})

		duplex.Link(sa, sb)
		a.Set(expected.key, expected.valueA)
	})

	t.Run("change in two-ways", func(t *testing.T) {
		a := NewSyncModel(sb.WithId("A"))
		b := NewSyncModel(sb.WithId("B"))

		sa := a.CreateStream(duplex.WithName("a->b"))
		sb := b.CreateStream(duplex.WithName("b->a"))

		a.Set(expected.key, expected.valueA)

		sb.On("changedByPeer", func(data interface{}) {
			assert.Equal(a.Id(), data.(*ValueModelFrom).From)
			assert.Equal(expected.key, data.(*ValueModelFrom).K)
			assert.Equal(expected.valueA, data.(*ValueModelFrom).V)
			assert.Equal(expected.valueA, b.Get(expected.key, false))

			a.On("changedByPeer", func(data interface{}) {
				assert.Equal(b.Id(), data.(*ValueModelFrom).From)
				assert.Equal(expected.key, data.(*ValueModelFrom).K)
				assert.Equal(expected.valueA, data.(*ValueModelFrom).V)
				assert.Equal(expected.valueA, a.Get(expected.key, false))
			})
		})

		duplex.Link(sa, sb)
	})
}
