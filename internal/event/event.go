package event

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Event struct {
	Data           EventData
	ForwardResults chan<- ForwardResult
}

func (e Event) SendForwardResult(ctx context.Context, subscriberId uuid.UUID, err error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	select {
	case <-ctx.Done():
	case e.ForwardResults <- ForwardResult{SubscriberId: subscriberId, Err: err}:
	}
}

type EventData struct {
	ClientAddr string
	Method     string
	Path       string
	Query      string
	Header     http.Header
	Body       []byte
}
