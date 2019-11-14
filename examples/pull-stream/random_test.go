package main

import (
	"testing"
)

var num = 500000

func BenchmarkRandomWrong(b *testing.B)  {
	for n := 0; n < b.N; n++ {
		loggerWrong(createRandomStream(num))
	}
}

func BenchmarkRandomFor(b *testing.B)  {
	for n := 0; n < b.N; n++ {
		loggerFor(createRandomStream(num))
	}
}

func BenchmarkRandomGo(b *testing.B)  {
	for n := 0; n < b.N; n++ {
		loggerGo(createRandomStream(num))
	}
	WG.Wait()
}
