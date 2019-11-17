package event

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestEmit(t *testing.T) {
	event := NewEvent()
	total1 := 0
	event.On("_update", func(data interface{}) {
		total1 = total1 + data.(int)
	})

	total2 := 0
	event.On("_update", func(data interface{}) {
		total2 = total2 + data.(int)
	})

	type user struct {
		Name string
		Age  int
	}
	counter := 0
	event.On("_user", func(data interface{}) {
		user := data.(user)
		assert.Equal(t, "Alice", user.Name)
		assert.Equal(t, 3, user.Age)
		counter++
	})

	event.Emit("_update", 1)
	event.Emit("_user", user{
		Name: "Alice",
		Age:  3,
	})
	event.Emit("_update", 2)
	event.Emit("_user", user{
		Name: "Alice",
		Age:  3,
	})

	assert.Equal(t, 3, total1)
	assert.Equal(t, 3, total2)
	assert.Equal(t, 2, counter)
}

func TestEvent_On(t *testing.T) {
	event := NewEvent()
	total := 0
	cb1 := func(data interface{}) {
		total = total + data.(int)
	}
	cb2 := func(data interface{}) {
		total = total + data.(int) + 1
	}
	event.On("_update", cb1)
	event.RemoveListener("_update", cb1)
	event.On("_update", cb2)
	event.Emit("_update", 1)
	assert.Equal(t, 2, total)
}

func TestEvent_RemoveListener(t *testing.T) {
	event := NewEvent()
	total := 0
	cb := func(data interface{}) {
		total = total + data.(int)
	}
	event.On("_update", cb)
	event.Emit("_update", 1)
	assert.Equal(t, 1, total)

	event.RemoveListener("_update", cb)
	event.Emit("_update", 1)
	assert.Equal(t, 1, total)
}

func TestEventRace(t *testing.T) {
	n := 100
	event := NewEvent()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for i := 0; i < n; i++ {
			event.Emit("_update", "_update")
			event.Emit("_update_once", "_update_once")
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < n; i++ {
			event.On("_update", func(data interface{}) {
			})
		}
		wg.Done()
	}()
	wg.Wait()
}
