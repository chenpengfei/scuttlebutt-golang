package pull_through

import (
	"errors"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/pull-stream/pkg/util"
)

type Event string

type Through struct {
	writer func(interface{})
	ender  func()
	queue  []interface{}
	ended  error
	err    error
	cb     pull.SourceCallback
}

func NewThrough(opts ...Option) *Through {
	t := &Through{
		writer: nil,
		ender:  nil,
		queue:  make([]interface{}, 0),
	}

	for _, opt := range opts {
		opt(t)
	}

	if t.writer == nil {
		t.writer = func(data interface{}) {
			t.EnQueue(data)
		}
	}

	if t.ender == nil {
		t.ender = func() {
			t.EnQueue(nil)
		}
	}

	return t
}

const (
	DataEvent Event = "data"
	EndEvent  Event = "end"
	ErrEvent  Event = "error"
)

func (t *Through) Emit(event Event, data interface{}) {
	if event == DataEvent {
		t.EnQueue(data)
	}
	if event == EndEvent {
		t.ended = pull.ErrPullStreamEnd
		t.EnQueue(nil)
	}
	if event == ErrEvent {
		if e, ok := data.(error); ok {
			t.err = e
		} else {
			t.err = errors.New("error")
		}
	}
}

func (t *Through) Through() pull.Through {
	return func(read pull.Read) pull.Read {
		return func(end error, cb pull.SourceCallback) {
			t.ended = util.ErrOr(t.ended, end)
			if end != nil {
				read(end, func(end error, data interface{}) {
					if t.cb != nil {
						_cb := t.cb
						t.cb = nil
						_cb(end, nil)
					}
					cb(end, nil)
				})
				return
			}

			t.cb = cb

			var next func()
			next = func() {
				//if it's an error
				if t.cb == nil {
					return
				}

				_cb := t.cb
				if t.err != nil {
					t.cb = nil
					_cb(t.err, nil)
				} else if len(t.queue) > 0 {
					data := t.queue[0]
					t.queue = t.queue[1:]
					t.cb = nil
					if data == nil {
						_cb(pull.ErrPullStreamEnd, data)
					} else {
						_cb(nil, data)
					}
				} else {
					read(t.ended, func(end error, data interface{}) {
						//null has no special meaning for pull-stream
						if end != nil && end != pull.ErrPullStreamEnd {
							t.err = end
							next()
							return
						}
						if t.ended = util.ErrOr(t.ended, end); t.ended != nil {
							t.ender()
						} else if data != nil {
							t.writer(data)
							if t.ended != nil || t.err != nil {
								read(util.ErrOr(t.err, t.ended), func(end error, data interface{}) {
									t.cb = nil
									_cb(util.ErrOr(t.err, t.ended), nil)
								})
								return
							}
						}
						next()
					})
				}
			}
			next()
		}
	}
}

func (t *Through) EnQueue(data interface{}) {
	t.queue = append(t.queue, data)
}
