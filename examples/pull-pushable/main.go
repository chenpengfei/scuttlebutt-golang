package main

import (
	"fmt"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	pullpushable "github.com/chenpengfei/scuttlebutt-golang/pkg/pull-pushable"
)

func logger(read pull.Read) {
	var next func(error, interface{})
	next = func(end error, data interface{}) {
		if end != nil {
			return
		}

		fmt.Println(data)
		read(nil, next)
	}
	read(nil, next)
}

func main() {
	source := pullpushable.NewPushable()

	source.Push(1)
	pull.Pull(source.Read, pull.Sink(logger))

	source.Push(2)
	source.Push(3)
	source.Ender(pull.ErrPullStreamEnd)
}
