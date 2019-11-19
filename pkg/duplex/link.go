package duplex

import "github.com/chenpengfei/pull-stream/pkg/pull"

func Link(a *Duplex, b *Duplex) {
	pull.Pull(a.GetSource(), b.GetSink())
	pull.Pull(b.GetSource(), a.GetSink())
}
