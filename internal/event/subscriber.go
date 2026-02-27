package event

import (
	"time"

	"github.com/google/uuid"
)

type Subscriber struct {
	Id       uuid.UUID
	JoinedAt time.Time
	Name     string
	Address  string
	Chan     chan<- Event
	Cancel   func()
}
