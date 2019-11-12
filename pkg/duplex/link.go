package duplex

import pullstream "scuttlebutt-golang/pkg/pull-stream"

func Link(a *Duplex, b *Duplex) {
	pullstream.Pull(a.GetSource(), b.GetSink())
	pullstream.Pull(b.GetSource(), a.GetSink())
}
