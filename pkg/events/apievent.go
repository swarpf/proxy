package events

import (
	"github.com/olebedev/emitter"
)

type ApiEventEmitter = emitter.Emitter

type ApiEventMsg struct {
	Request  string
	Response string
}
