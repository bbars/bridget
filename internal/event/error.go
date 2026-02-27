package event

import (
	"github.com/google/uuid"
)

type ForwardResult struct {
	Err          error
	SubscriberId uuid.UUID
}
