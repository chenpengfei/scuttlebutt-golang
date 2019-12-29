package model

import (
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestModelFilter(t *testing.T) {
	assert := assert.New(t)

	accept := &scuttlebutt.Accept{
		Blacklist: nil,
		Whitelist: []string{"foo"},
	}

	expected := struct {
		key    string
		valueA string
		valueB string
	}{
		key:    "foo",
		valueA: "changed by A",
		valueB: "changed by B",
	}

	ignored := struct {
		key    string
		valueA string
	}{
		key:    "ignored",
		valueA: "changed by A",
	}

	t.Run("whitelist-filter out in history", func(t *testing.T) {
		a := NewSyncModel(scuttlebutt.WithId("A"))
		b := NewSyncModel(scuttlebutt.WithId("B"), scuttlebutt.WithAccept(accept))

		sa := a.CreateStream()
		sb := b.CreateStream()

		a.Set(expected.key, expected.valueA)
		//todo.时间算法防止冲突，去掉所有单元测试中的 sleep
		time.Sleep(time.Second)
		a.Set(ignored.key, expected.valueA)

		sb.On("synced", func(data interface{}) {
			assert.Equal(expected.valueA, b.Get(expected.key, false))
			assert.Equal(nil, b.Get(ignored.key, false))
		})

		duplex.Link(sa, sb)
	})

	t.Run("whitelist-filter out in following update", func(t *testing.T) {
		a := NewSyncModel(scuttlebutt.WithId("A"))
		b := NewSyncModel(scuttlebutt.WithId("B"), scuttlebutt.WithAccept(accept))

		sa := a.CreateStream()
		sb := b.CreateStream()

		duplex.Link(sa, sb)

		a.Set(expected.key, expected.valueA)
		time.Sleep(time.Second)
		a.Set(ignored.key, expected.valueA)

		assert.Equal(expected.valueA, a.Get(expected.key, false))
		assert.Equal(expected.valueA, a.Get(ignored.key, false))

		assert.Equal(expected.valueA, b.Get(expected.key, false))
		assert.Equal(nil, b.Get(ignored.key, false))
	})
}
