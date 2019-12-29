package scuttlebutt

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type CreateIdFn func() SourceId

func CreateId() SourceId {
	b := make([]byte, 16)
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	r.Read(b)

	return SourceId(fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
}

func Filter(update *Update, sources Sources) bool {
	// Update in local store
	if sources == nil {
		return true
	}
	v, found := sources[update.SourceId]
	if !found {
		return true
	}
	if update.Timestamp > v {
		return true
	}
	return false
}

func Sort(history []*Update) {
	// sort by timestamps, then ids.
	// there should never be a pair with equal timestamps
	// and ids.
	sort.SliceStable(history, func(i, j int) bool {
		if history[i].Timestamp < history[j].Timestamp {
			return true
		}
		if history[i].Timestamp > history[j].Timestamp {
			return false
		}
		return history[i].SourceId < history[j].SourceId
	})
}
