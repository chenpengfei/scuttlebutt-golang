package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"scuttlebutt-golang/pkg/model"
)

func InitLogWriterEarly() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	formatter := log.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
	}
	log.SetFormatter(&formatter)
}

func PrintKeyValue(model *model.SyncModel, key string) {
	log.WithField("id", model.Id()).WithField("value", model.Get(key, false)).Info("With clock")
	log.WithField("id", model.Id()).WithField("value", model.Get(key, true)).Info("With no clock")

}
