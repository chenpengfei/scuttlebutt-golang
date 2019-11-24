package test

import (
	"errors"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/model"
	sb "github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReadable(t *testing.T) {
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

	t.Run("A is read-ony to B (changed before sync)", func(t *testing.T) {
		a := model.NewSyncModel(sb.WithId("A"))
		b := model.NewSyncModel(sb.WithId("B"))

		sa := a.CreateReadStream(duplex.WithName("a->b"))
		sb := b.CreateStream(duplex.WithName("b->a"))

		a.Set(expected.key, expected.valueA)

		sb.On("synced", func(data interface{}) {
			// A won't be changed by B
			assert.Error(errors.New("A won't be changed by B"))
		})

		duplex.Link(sa, sb)
	})

	t.Run("A is read-ony to B (changed after sync)", func(t *testing.T) {
		a := model.NewSyncModel(sb.WithId("A"))
		b := model.NewSyncModel(sb.WithId("B"))

		sa := a.CreateReadStream(duplex.WithName("a->b"))
		sb := b.CreateStream(duplex.WithName("b->a"))

		a.Set(expected.key, expected.valueA)

		duplex.Link(sa, sb)

		time.Sleep(10 * time.Millisecond)
		b.Set(expected.key, expected.valueB)

		assert.Equal(a.Get(expected.key, false), expected.valueA)
		assert.Equal(b.Get(expected.key, false), expected.valueB)
	})

	t.Run("B is write-only to A (changed before sync)", func(t *testing.T) {
		a := model.NewSyncModel(sb.WithId("A"))
		b := model.NewSyncModel(sb.WithId("B"))

		sa := a.CreateStream(duplex.WithName("a->b"))
		sb := b.CreateWriteStream(duplex.WithName("b->a"))

		b.Set(expected.key, expected.valueB)

		duplex.Link(sa, sb)

		assert.Equal(nil, a.Get(expected.key, false))
	})

	t.Run("B is write-only to A (changed after sync)", func(t *testing.T) {
		a := model.NewSyncModel(sb.WithId("A"))
		b := model.NewSyncModel(sb.WithId("B"))

		sa := a.CreateStream(duplex.WithName("a->b"))
		sb := b.CreateWriteStream(duplex.WithName("b->a"))

		duplex.Link(sa, sb)

		a.Set(expected.key, expected.valueA)
		time.Sleep(10 * time.Millisecond)
		b.Set(expected.key, expected.valueB)

		assert.Equal(a.Get(expected.key, false), expected.valueA)
		assert.Equal(b.Get(expected.key, false), expected.valueB)
	})

	t.Run("A is read-ony and B is write-only (changed after sync)", func(t *testing.T) {
		a := model.NewSyncModel(sb.WithId("A"))
		b := model.NewSyncModel(sb.WithId("B"))

		sa := a.CreateReadStream(duplex.WithName("a->b"))
		sb := b.CreateWriteStream(duplex.WithName("b->a"))

		duplex.Link(sa, sb)

		a.Set(expected.key, expected.valueA)
		assert.Equal(b.Get(expected.key, false), expected.valueA)

		time.Sleep(10 * time.Millisecond)
		b.Set(expected.key, expected.valueB)

		assert.Equal(a.Get(expected.key, false), expected.valueA)
		assert.Equal(b.Get(expected.key, false), expected.valueB)
	})
}
