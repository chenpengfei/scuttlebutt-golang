package socket

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtils(t *testing.T) {
	assert := assert.New(t)

	t.Run("Serialize then Parse", func(t *testing.T) {
		msg := "hello"
		length, payload := Parse(Serialize([]byte(msg)))
		assert.Equal(len(msg), length)
		assert.Equal(msg, string(payload))
	})
}
