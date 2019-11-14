package main

import (
	"math/rand"
	pullstream "scuttlebutt-golang/pkg/pull-stream"
	"strconv"
	"sync"
)

// random is a source n of random numbers.
func createRandomStream(n int) pullstream.Read {
	return func(end pullstream.EndOrError, cb pullstream.SourceCallback) {
		if end.Yes() {
			cb(end, nil)
			return
		}

		// only read n times, then stop.
		n--
		if n < 0 {
			cb(pullstream.End, nil)
			return
		}

		cb(pullstream.Null, rand.Intn(100))
	}
}

// always record the result to prevent
// the compiler eliminating the function call.
var result string

func dealData(n interface{}) string {
	if d, ok := n.(int); ok {
		return strconv.Itoa((d*d*d + d) / (d + 1))
	} else {
		return ""
	}
}

func loggerWrong(read pullstream.Read) {
	var next func(pullstream.EndOrError, interface{})
	next = func(end pullstream.EndOrError, data interface{}) {
		if end.Yes() {
			return
		}

		result = dealData(data)
		read(pullstream.Null, next)
	}
	read(pullstream.Null, next)
}

var WG sync.WaitGroup

// logger reads a source and logs it.
func loggerGo(read pullstream.Read) {
	WG.Add(1)
	var next func(pullstream.EndOrError, interface{})
	next = func(end pullstream.EndOrError, data interface{}) {
		if end.Yes() {
			WG.Done()
			return
		}

		result = dealData(data)
		go read(pullstream.Null, next)
	}
	read(pullstream.Null, next)
}

func loggerForChannel(read pullstream.Read) {
	c := make(chan struct{}, 1)
	c <- struct{}{}
	for {
		select {
		case _, ok := <-c:
			if !ok {
				return
			}
			read(pullstream.Null, func(end pullstream.EndOrError, data interface{}) {
				if end.Yes() {
					close(c)
					return
				}

				result = dealData(data)
				c <- struct{}{}
			})
		}
	}
}

func loggerForWaitGroup(read pullstream.Read) {
	var wg sync.WaitGroup
	over := false
	for {
		wg.Add(1)
		read(pullstream.Null, func(end pullstream.EndOrError, data interface{}) {
			if end.Yes() {
				wg.Done()
				over = true
				return
			}

			result = dealData(data)
			wg.Done()
		})
		wg.Wait()
		if over {
			break
		}
	}
}
