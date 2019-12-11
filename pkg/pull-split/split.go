package pull_split

import (
	"github.com/chenpengfei/pull-stream/pkg/pull"
	pullthrough "github.com/chenpengfei/scuttlebutt-golang/pkg/pull-through"
	"strings"
)

type Split struct {
	matcher string
	mapper  func(interface{}) interface{}
	reverse bool
	last    bool
	soFar   string

	through *pullthrough.Through
}

func NewSplit(opts ...Option) *Split {
	s := &Split{
		matcher: "\n",
		mapper:  nil,
		reverse: false,
		last:    false,
		soFar:   "",
	}

	s.through = pullthrough.NewThrough(
		pullthrough.WithWriter(s.writer),
		pullthrough.WithEnder(s.ender))

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Split) enqueue(piece interface{}) {
	if s.mapper != nil {
		piece = s.mapper(piece)
	}
	s.through.EnQueue(piece)
}

func (s *Split) writer(buffer interface{}) {
	//todo
	buf := buffer.(string)

	if s.reverse {
		buf = buf + s.soFar
	} else {
		buf = s.soFar + buf
	}
	pieces := strings.Split(buf, s.matcher)
	if s.reverse {
		s.soFar = pieces[0]
		pieces = pieces[1:]
	} else {
		s.soFar = pieces[len(pieces)-1]
		pieces = pieces[:len(pieces)-1]
	}
	l := len(pieces)
	var data string
	for i := 0; i < l; i++ {
		if s.reverse {
			data = pieces[l-1-i]
		} else {
			data = pieces[i]
		}
		//todo
		if data == "" {
			continue
		}
		s.enqueue(data)
	}
}

func (s *Split) ender() {
	if s.last {
		if s.soFar == "" {
			s.enqueue(nil)
		} else {
			s.enqueue(s.soFar)
		}
		return
	}

	s.enqueue(nil)
}

func (s *Split) Through() pull.Through {
	return s.through.Through()
}
