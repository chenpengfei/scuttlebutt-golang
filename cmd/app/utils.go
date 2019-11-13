package main

import (
	log "github.com/sirupsen/logrus"
	"scuttlebutt-golang/pkg/model"
)

func PrintKeyValue(model *model.SyncModel, key string) {
	log.WithField("id", model.Id()).WithField("value", model.Get(key, false)).Info("With clock")
	log.WithField("id", model.Id()).WithField("value", model.Get(key, true)).Info("With no clock")

}
