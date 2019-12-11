package pull_split

type Option func(split *Split)

func WithMatcher(matcher string) Option {
	return func(split *Split) {
		split.matcher = matcher
	}
}

func WithMapper(mapper func(interface{}) interface{}) Option {
	return func(split *Split) {
		split.mapper = mapper
	}
}

func WithReverse(reverse bool) Option {
	return func(split *Split) {
		split.reverse = reverse
	}
}

func WithLast(last bool) Option {
	return func(split *Split) {
		split.last = last
	}
}
