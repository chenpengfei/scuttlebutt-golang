package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	sb "scuttlebutt-golang/pkg"
	"time"
)

func main() {
	initLogWriterEarly()
	ctx := context.Background()

	cs := sb.NewStream(ctx, sb.Scuttlebutt{
		Protocol: sb.NewModel("Client"),
		Node: sb.Node{
			ID:        "XXXX",
			Timestamp: time.Now(),
		},
	})

	ss := sb.NewStream(ctx, sb.Scuttlebutt{
		Protocol: sb.NewModel("Server"),
		Node: sb.Node{
			ID:        "YYYY",
			Timestamp: time.Now(),
		},
	})

	cs.Link(ss)
	ss.Link(cs)

	counter := 0
	for {
		counter++
		ternarys := make([]*sb.Ternary, 1)
		ternarys[0] = &sb.Ternary{
			Key:       sb.KeySpeed,
			Value:     counter,
			Timestamp: time.Now(),
		}
		cs.ApplyUpdates(&sb.Update{Ternarys: ternarys})
		time.Sleep(time.Second)
	}
}

func initLogWriterEarly() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	formatter := log.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
	}
	log.SetFormatter(&formatter)
}
