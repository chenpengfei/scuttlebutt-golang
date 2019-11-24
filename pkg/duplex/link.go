package duplex

import "github.com/chenpengfei/pull-stream/pkg/pull"

func Link(a *Duplex, b *Duplex) {
	pull.Pull(a.Source(), b.Sink())
	pull.Pull(b.Source(), a.Sink())
}
