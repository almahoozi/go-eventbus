package eventbus

import (
	"time"

	"github.com/almahoozi/go-eventbus/pkg/id"
)

type (
	Stringer interface {
		String() string
	}
	Event struct {
		ID             string      `json:"id"`
		Name           Stringer    `json:"name"`
		Data           interface{} `json:"data"`
		Timestamp      time.Time   `json:"timestamp"`
		handlerTimeout time.Duration
		publishTimeout time.Duration
	}
)

func newEvent(name Stringer, data interface{}) Event {
	return Event{
		ID:        id.New(),
		Name:      name,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}
