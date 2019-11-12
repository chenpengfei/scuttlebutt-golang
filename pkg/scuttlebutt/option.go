package scuttlebutt

type Option func(*Scuttlebutt)

func WithId(id SourceId) Option {
	return func(scuttlebutt *Scuttlebutt) {
		scuttlebutt.Id = id
	}
}
