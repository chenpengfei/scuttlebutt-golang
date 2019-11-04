package pkg

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Value struct {
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type Model struct {
	Name string
	storeMap map[string]Value
}

func NewModel(name string) *Model {
	return &Model{
		Name: name,
		storeMap: make(map[string]Value),
	}
}

func (m *Model) ApplyUpdates(update *Update) {
	for _, new := range update.Ternarys {
		if old, ok := m.storeMap[new.Key]; ok {
			if new.Timestamp.Before(old.Timestamp) {
				continue
			}
		}
		m.storeMap[new.Key] = Value{
			Value:     new.Value,
			Timestamp: new.Timestamp,
		}
		log.WithField("node_name", m.Name).
			WithField("key", new.Key).
			WithField("value", new.Value).
			WithField("timestamp", new.Timestamp.Format("2006/01/02 15:04:05.000")).
			Debug("apply update.")
	}
}

func (m *Model) History(peerClock time.Time) *Update {
	ternarys := make([]*Ternary, 0)
	for k, v := range m.storeMap {
		if v.Timestamp.After(peerClock) {
			ternarys = append(ternarys, &Ternary{
				Key:       k,
				Value:     v.Value,
				Timestamp: v.Timestamp,
			})
		}
	}
	return &Update{Ternarys: ternarys}
}
