package pull_through

type Option func(through *Through)

func WithWriter(writer func(interface{})) Option {
	return func(through *Through) {
		through.writer = writer
	}
}

func WithEnder(ender func()) Option {
	return func(through *Through) {
		through.ender = ender
	}
}
