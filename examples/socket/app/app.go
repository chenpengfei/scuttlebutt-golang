package main

import (
	"context"
	cw "github.com/chenpengfei/context-wrapper"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/model"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/socket"
	log "github.com/sirupsen/logrus"
	"net"
)

func main() {
	ctx := cw.WithSignal(context.Background())

	address := "localhost:9989"

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.WithError(err).Error("connect to cloud failed")
		return
	}

	socket := socket.NewDuplex(conn, nil)

	log.WithField("address", address).Info("connected to cloud")

	signalModel := model.NewSyncModel()
	app2c := signalModel.CreateSinkStream()
	app2c.On("synced", func(data interface{}) {
		log.Info("signal model has synced with cloud")
	})

	duplex.Link(socket, app2c)

	signalModel.On("changedByPeer", func(data interface{}) {
		speed := signalModel.Get("speed", false)
		log.WithField("speed", speed).Info("changedByPeer")
	})
	signalModel.On("error", func(data interface{}) {
		log.Error("connection has broken")
	})

	<-ctx.Done()

	log.Info("I have to go...")

	// send some raw data to server
	// echo -n "test out the server" | nc localhost 8080
}
