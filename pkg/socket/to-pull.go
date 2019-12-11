package socket

import (
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"io"
)

func (d *Duplex) rawSource() pull.Read {
	var ended error
	did := false
	cbChan := make(chan pull.SourceCallback, 1)

	done := func() {
		if did {
			return
		}
		did = true

		close(cbChan)
		_ = d.stream.Close()
	}

	go func() {
		buf := make([]byte, 32*1024)

		for {
			cb, ok := <-cbChan
			if !ok {
				return
			}

			nr, err := d.stream.Read(buf)
			if err != nil {
				ended = err
				if ended == io.EOF {
					ended = pull.ErrPullStreamEnd
				}
			}

			if ended != nil {
				done()
			}
			//todo
			cb(ended, string(buf[0:nr]))
		}
	}()

	return func(abort error, cb pull.SourceCallback) {
		if cb == nil {
			panic("*must* provide cb")
		}

		if abort != nil {
			done()
			cb(abort, nil)
			return
		}

		if ended != nil {
			done()
			cb(ended, nil)
			return
		}

		cbChan <- cb
	}
}

func (d *Duplex) rawSink(read pull.Read) {
	nextChan := make(chan struct{}, 1)
	nextChan <- struct{}{}
	var ended error
	did := false

	done := func(end error) {
		if did {
			return
		}
		did = true

		close(nextChan)
		_ = d.stream.Close()
		if d.cb != nil {
			d.cb(end, nil)
		}
	}

	go func() {
		for {
			_, ok := <-nextChan
			if !ok {
				return
			}
			read(ended, func(end error, data interface{}) {
				if end != nil {
					done(end)
					return
				}

				if ended != nil {
					done(ended)
					return
				}

				n, err := d.stream.Write([]byte(data.(string)))
				if err != nil {
					ended = err
					if ended == io.EOF {
						ended = pull.ErrPullStreamEnd
					}
				}
				if n != len(data.(string)) {
					ended = io.ErrShortWrite
				}

				nextChan <- struct{}{}
			})
		}
	}()
}

func (d *Duplex) Source() interface{} {
	return d.rawSource()
}

func (d *Duplex) Sink() interface{} {
	return pull.Sink(d.rawSink)
}

type Duplex struct {
	stream io.ReadWriteCloser
	cb     pull.SourceCallback
}

func NewDuplex(stream io.ReadWriteCloser, cb pull.SourceCallback) *Duplex {
	return &Duplex{
		stream: stream,
		cb:     cb,
	}
}
