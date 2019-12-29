package pull_stringify

import (
	"bytes"
	"encoding/json"
	"github.com/chenpengfei/pull-stream/pkg/pull"
)

type PullStringify struct {
	open   string
	prefix string
	suffix string
	close  string
	indent string

	//todo
	buf bytes.Buffer
	enc *json.Encoder
}

func NewPullStringify(opts ...Option) *PullStringify {
	// default is pretty double newline delimited json
	s := &PullStringify{
		open:   "",
		prefix: "",
		suffix: "\n\n",
		close:  "",
		indent: "  ",
	}

	for _, opt := range opts {
		opt(s)
	}

	s.enc = json.NewEncoder(&s.buf)
	s.enc.SetIndent(s.prefix, s.indent)

	return s
}

func (s *PullStringify) Serialize() pull.Through {
	var ended error
	first := true

	return func(read pull.Read) pull.Read {
		return func(end error, cb pull.SourceCallback) {
			if ended != nil {
				cb(ended, nil)
			}
			if end != nil {
				cb(end, nil)
			}

			read(nil, func(end error, data interface{}) {
				if end != nil {
					ended = end
					if ended != pull.ErrPullStreamEnd {
						cb(ended, nil)
						return
					}
					data := s.close
					if first {
						data = s.open + s.close
					}
					cb(nil, data)
				} else {
					f := first
					first = false

					err := s.enc.Encode(data)
					if err != nil {
						ended = err
						cb(err, nil)
					} else {
						prefix := s.prefix
						if f {
							prefix = s.open
						}
						val := prefix + s.buf.String() + s.suffix
						s.buf.Reset()
						cb(nil, val)
					}
				}
			})
		}
	}
}
