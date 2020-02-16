package scuttlebutt

type Option func(*Scuttlebutt)

func WithId(id SourceId) Option {
	return func(scuttlebutt *Scuttlebutt) {
		scuttlebutt.id = id
	}
}

func WithAccept(accept *ModelAccept) Option {
	return func(scuttlebutt *Scuttlebutt) {
		scuttlebutt.Accept = accept
	}
}
