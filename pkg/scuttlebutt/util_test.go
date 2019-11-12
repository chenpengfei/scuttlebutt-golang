package scuttlebutt

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestCreateId(t *testing.T) {
	id := CreateId()
	assert.Equal(t, 36, len(id))
	assert.Equal(t, 4, strings.Count(string(id), "-"))
}

func TestFilter(t *testing.T) {
	u1 := &Update{
		SourceId:  "XX",
		Timestamp: 1,
	}
	u21 := &Update{
		SourceId:  "YY",
		Timestamp: 1,
	}
	u22 := &Update{
		SourceId:  "YY",
		Timestamp: 2,
	}

	assert.True(t, Filter(u1, nil))

	s1 := &Sources{Map: nil}
	assert.True(t, Filter(u1, s1))

	s2 := &Sources{Map: make(map[SourceId]Timestamp)}
	s2.Map[u21.SourceId] = u21.Timestamp
	assert.True(t, Filter(u1, s2))
	assert.True(t, Filter(u22, s2))
}

func TestSort(t *testing.T) {
	u1 := &Update{
		SourceId:  "XX",
		Timestamp: 1,
	}
	u21 := &Update{
		SourceId:  "YY",
		Timestamp: 1,
	}
	u22 := &Update{
		SourceId:  "YY",
		Timestamp: 2,
	}

	h := make([]*Update, 3)
	h[0] = u22
	h[1] = u21
	h[2] = u1
	Sort(h)
	assert.Equal(t, u1, h[0])
	assert.Equal(t, u21, h[1])
	assert.Equal(t, u22, h[2])
}

type Any struct {
	Name  string
	Speed int
}

func TestUpdate_Sign(t *testing.T) {
	u := &Update{
		SourceId:  "XX",
		Timestamp: Timestamp(time.Now().Second()),
		Data: Any{
			Name:  "lixiang",
			Speed: 120,
		},
	}

	digest, err := Sha256Sign(u)
	assert.Nil(t, err)
	u.Digest = digest
	assert.True(t, Sha256Verify(u))
	u.Digest = "digest"
	assert.False(t, Sha256Verify(u))
}
