package main

import (
	"context"
	cw "github.com/chenpengfei/context-wrapper"
	reconnect "github.com/chenpengfei/reconnect-core"
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

	rc := reconnect.NewReconnection(ctx)
	rc.OnConnect(func(conn net.Conn) {
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
	})
	rc.OnNotify(func(err error, duration time.Duration) {
		log.WithError(err).WithField("next", duration).Error("retry...")
	})
	rc.OnError(func(err error) {
		log.WithError(err).Error("connection has broken")
	})
	rc.Dial("tcp", address)

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
