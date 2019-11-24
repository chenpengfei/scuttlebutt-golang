package duplex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuffer(t *testing.T) {
	assert := assert.New(t)

	expected := struct {
		first  string
		second string
	}{
		first:  "first",
		second: "second",
	}

	t.Run("push then shift", func(t *testing.T) {
		buffer := NewBuffer()
		buffer.Push(expected.first)
		buffer.Push(expected.second)
		assert.Equal(expected.first, buffer.Shift())
		assert.Equal(expected.second, buffer.Shift())
	})

	t.Run("unshift then shift", func(t *testing.T) {
		buffer := NewBuffer()
		buffer.Unshift(expected.first)
		buffer.Unshift(expected.second)
		assert.Equal(expected.second, buffer.Shift())
		assert.Equal(expected.first, buffer.Shift())
	})

	t.Run("shift empty buffer", func(t *testing.T) {
		buffer := NewBuffer()
		assert.Equal(nil, buffer.Shift())
	})

	t.Run("length", func(t *testing.T) {
		buffer := NewBuffer()
		buffer.Push(expected.first)
		buffer.Unshift(expected.second)
		buffer.Shift()
		assert.Equal(1, buffer.Length())
	})
}
