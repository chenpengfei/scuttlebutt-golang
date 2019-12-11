package pull_through

import (
	"errors"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/pull-stream/pkg/sinks"
	"github.com/chenpengfei/pull-stream/pkg/sources"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestThrough(t *testing.T) {
	assert := assert.New(t)

	t.Run("emit error", func(t *testing.T) {
		_err := errors.New("expected error")

		var through *Through
		through = NewThrough(WithWriter(func(data interface{}) {
			through.EnQueue(_err)
		}))
		drain, _ := sinks.Drain(nil, func(end error, data interface{}) {
			assert.Error(_err, end)
		})

		pull.Pull(
			sources.Values([]int{1, 2, 3}, nil),
			through.Through(),
			drain)
	})

	t.Run("abort source on end within writer", func(t *testing.T) {
		_err := errors.New("intentional")

		var through *Through
		through = NewThrough(WithWriter(func(data interface{}) {
			//do nothing. this will make through read ahead some more.
			through.Emit(EndEvent, _err)
		}))
		drain, _ := sinks.Drain(nil, func(end error, data interface{}) {
			assert.Equal(end, pull.ErrPullStreamEnd)
		})

		pull.Pull(sources.Values([]int{1, 2, 3}, nil),
			through.Through(),
			drain)
	})
}
