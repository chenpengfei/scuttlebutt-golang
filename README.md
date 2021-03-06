# Scuttlebutt Golang 
[![Build Status](https://travis-ci.com/chenpengfei/scuttlebutt-golang.svg?branch=master)](https://travis-ci.com/chenpengfei/scuttlebutt-golang)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/chenpengfei/scuttlebutt-golang/blob/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/github/chenpengfei/scuttlebutt-golang/badge.svg)](https://coveralls.io/github/chenpengfei/scuttlebutt-golang)
[![Go Report Card](https://goreportcard.com/badge/github.com/chenpengfei/scuttlebutt-golang)](https://goreportcard.com/report/github.com/chenpengfei/scuttlebutt-golang)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)
[![GoDoc](https://godoc.org/github.com/chenpengfei/scuttlebutt-golang?status.svg)](https://godoc.org/github.com/chenpengfei/scuttlebutt-golang)

尝试解决两个一致性问题
1. 最终一致性问题中的 “最终结果同步”，用一个分布式 "HashMap" 解决
2. “过程完全回放” 问题，用一个分布式 "Queue" 解决

## Usage

```
a := model.NewSyncModel(sb.WithId("A"))
b := model.NewSyncModel(sb.WithId("B"))

sa := a.CreateStream(duplex.WithName("a->b"))
sb := b.CreateStream(duplex.WithName("b->a"))

a.Set("foo", "changed by A")

sb.On("synced", func(data interface{}) {
    PrintKeyValue(b, "foo")
})

duplex.Link(sa, sb)
```

## Run
```
make run
```

