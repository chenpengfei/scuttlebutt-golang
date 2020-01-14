package pull_pushable

type Option func(*Pushable)

func WithName(name string) Option {
	return func(pushable *Pushable) {
		pushable.name = name
	}
}

func WithOnClose(onClose OnClose) Option {
	return func(pushable *Pushable) {
		pushable.onClose = onClose
	}
}
