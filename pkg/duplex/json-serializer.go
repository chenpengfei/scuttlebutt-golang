package duplex

import (
	"encoding/json"
	"github.com/chenpengfei/pull-stream/pkg/pull"
	"github.com/chenpengfei/pull-stream/pkg/throughs"
	pullsplit "github.com/chenpengfei/scuttlebutt-golang/pkg/pull-split"
	ps "github.com/chenpengfei/scuttlebutt-golang/pkg/pull-stringify"
	"github.com/chenpengfei/scuttlebutt-golang/pkg/scuttlebutt"
)

func Serialize() pull.Through {
	ps := ps.NewPullStringify(
		ps.WithSuffix("\n"),
		ps.WithIndent(""),
	)

	return ps.Serialize()
}

func Parse() pull.Pulls {
	split := pullsplit.NewSplit()
	jsonParse := throughs.Map(func(in interface{}) (interface{}, error) {
		outgoing := &Outgoing{}
		update := &scuttlebutt.Update{}
		var sync string

		//todo
		data := in.(string)
		err := json.Unmarshal([]byte(data), &update)
		if err == nil && update.SourceId != "" {
			return update, nil
		}

		err = json.Unmarshal([]byte(data), &outgoing)
		if err == nil && outgoing.Id != "" {
			return outgoing, nil
		}

		err = json.Unmarshal([]byte(data), &sync)
		return sync, err
	})
	return pull.Pull(split.Through(), jsonParse).(pull.Pulls)
}
