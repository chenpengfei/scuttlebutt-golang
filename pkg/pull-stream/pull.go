package pull_stream

type EndOrError string

func (e EndOrError) Yes() bool {
	if e == End || e == Err {
		return true
	} else {
		return false
	}
}

func (e EndOrError) Error() bool {
	if e == Err {
		return true
	} else {
		return false
	}
}

func (e EndOrError) End() bool {
	if e == End {
		return true
	} else {
		return false
	}
}

const (
	End  EndOrError = "END"
	Err  EndOrError = "ERR"
	Null EndOrError = "NULL"
)

type SourceCallback func(end EndOrError, data interface{})
type SourceCallbackMap func(cb SourceCallback) SourceCallback
type Read func(end EndOrError, cb SourceCallback)
type Sink func(read Read)
type Through func(read Read) Read

func Pull(opts ...interface{}) interface{} {
	len := len(opts)
	if len == 0 {
		return nil
	}
	if len == 1 {
		return opts[0]
	}

	var read Read
	for i := 0; i < len; i++ {
		if s, ok := opts[i].(Read); ok {
			if i != 0 {
				panic("partial source must be at the begin of pipeline!")
			}
			read = s
			continue
		}
		if s, ok := opts[i].(Through); ok {
			read = s(read)
			continue
		}
		if s, ok := opts[i].(Sink); ok {
			if i != len-1 {
				panic("partial sink must be at the end of pipeline!")
			}
			s(read)
			return nil
		}
		panic("unknown pull-stream type")
	}
	return read
}
