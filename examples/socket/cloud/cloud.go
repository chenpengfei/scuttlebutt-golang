package main

import (
	"context"
	cw "github.com/chenpengfei/context-wrapper"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/duplex"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/logger"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/model"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/socket"
	log "github.com/sirupsen/logrus"
	"net"
)

func main() {
	ctx := cw.WithSignal(context.Background())

	signalModel := model.NewSyncModel()

	// vehicle
	go func() {
		address := "localhost:9988"

		vLog := logger.WithNamespace("veh")

		listener, err := net.Listen("tcp", address)
		if err != nil {
			vLog.WithError(err).Error("start vehicle server failed")
			return
		}
		defer listener.Close()

		log.WithField("port", address).Info("server to vehicle is listening")

		for {
			conn, err := listener.Accept()
			if err != nil {
				vLog.WithError(err).Error("accept connection failed")
				return
			}

			vLog.Info("one of vehicle has connected")

			socket := socket.NewDuplex(conn, nil)

			c2veh := signalModel.CreateSinkStream()
			c2veh.On("synced", func(data interface{}) {
				vLog.Info("signal model has synced with vehicle")
			})

			go duplex.Link(socket, c2veh)
		}
	}()

	// app
	go func() {
		address := "localhost:9989"

		appLog := logger.WithNamespace("app")

		listener, err := net.Listen("tcp", address)
		if err != nil {
			appLog.WithError(err).Error("start app server failed")
			return
		}
		defer listener.Close()

		log.WithField("port", address).Info("server to app is listening")

		for {
			conn, err := listener.Accept()
			if err != nil {
				appLog.WithError(err).Error("accept connection failed")
				return
			}

			appLog.Info("one of app has connected")

			socket := socket.NewDuplex(conn, nil)

			//c2app := signalModel.CreateSourceStream()
			c2app := signalModel.CreateStream()
			c2app.On("synced", func(data interface{}) {
				appLog.Info("signal model has synced with app")
			})

			go duplex.Link(socket, c2app)
		}
	}()

	<-ctx.Done()

	log.Info("I have to go...")
}
