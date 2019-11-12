package duplex

type Option func(s *Duplex)

func WithName(name string) Option {
	return func(s *Duplex) {
		s.name = name
	}
}

func WithReadable(readable bool) Option {
	return func(s *Duplex) {
		s.readable = readable
	}
}

func WithWritable(writable bool) Option {
	return func(s *Duplex) {
		s.writable = writable
	}
}
