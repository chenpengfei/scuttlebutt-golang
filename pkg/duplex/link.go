package duplex

import "github.com/chenpengfei/pull-stream/pkg/pull"

func Link(a Stream, b Stream) {
	pull.Pull(a.Source(), b.Sink())
	pull.Pull(b.Source(), a.Sink())
}
