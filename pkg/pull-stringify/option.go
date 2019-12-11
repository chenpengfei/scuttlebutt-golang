package pull_stringify

type Option func(stringify *PullStringify)

func WithOpen(open string) Option {
	return func(stringify *PullStringify) {
		stringify.open = open
	}
}

func WithPrefix(prefix string) Option {
	return func(stringify *PullStringify) {
		stringify.prefix = prefix
	}
}

func WithSuffix(suffix string) Option {
	return func(stringify *PullStringify) {
		stringify.suffix = suffix
	}
}

func WithClose(close string) Option {
	return func(stringify *PullStringify) {
		stringify.close = close
	}
}

func WithIndent(indent string) Option {
	return func(stringify *PullStringify) {
		stringify.indent = indent
	}
}
