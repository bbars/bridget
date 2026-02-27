package forwarder

import (
	"github.com/bbars/bridget/internal/event"
	"github.com/google/uuid"
)

type Subscriptions interface {
	Load(requestPath string) (event.Subscribers, bool)
	Store(pattern string, subr event.Subscriber) error
	Delete(pattern string, subrId uuid.UUID)
}
