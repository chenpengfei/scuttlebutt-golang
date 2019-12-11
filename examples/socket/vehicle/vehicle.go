package main

import (
	"context"
	cw "github.com/chenpengfei/context-wrapper"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/model"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/socket"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

func main() {
	ctx := cw.WithSignal(context.Background())

	signalModel := model.NewSyncModel()

	address := "localhost:9988"

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.WithError(err).Error("connect to cloud failed")
		return
	}

	socket := socket.NewDuplex(conn, nil)

	log.WithField("address", address).Info("connected to cloud")

	veh2c := signalModel.CreateSourceStream()
	veh2c.On("synced", func(data interface{}) {
		log.Info("signal model has synced with cloud")
	})

	duplex.Link(socket, veh2c)

	signalModel.On("changedByPeer", func(data interface{}) {
		speed := signalModel.Get("speed", false)
		log.WithField("speed", speed).Info("changedByPeer")
	})
	signalModel.On("error", func(data interface{}) {
		log.Error("connection has broken")
	})

	go func() {
		speed := 0
		for {
			signalModel.Set("speed", speed)
			speed++
			time.Sleep(time.Second)
		}
	}()

	<-ctx.Done()

	log.Info("I have to go...")
}
