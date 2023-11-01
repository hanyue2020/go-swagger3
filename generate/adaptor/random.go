package adaptor

import (
	"math/rand"
	"time"

	"github.com/spf13/cast"
)

type PickUp struct{}

func (m *PickUp) Gen(args ...string) any {
	if len(args) == 0 {
		return ""
	}
	index := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(args))
	return args[index]
}

type RandomInt struct{}

func (m *RandomInt) Gen(args ...string) any {
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch len(args) {
	case 1:
		return rander.Int63n(cast.ToInt64(args[0]))
	case 2:
		l := cast.ToInt64(args[0])
		r := cast.ToInt64(args[1])
		diff := r - l
		return l + rander.Int63n(diff)
	default:
		return rander.Int63()
	}
}
