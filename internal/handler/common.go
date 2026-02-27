package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bbars/bridget/internal/event"
	"github.com/google/uuid"
)

func getRequestedSubscriberName(r *http.Request) string {
	return r.URL.Query().Get("name")
}

func requestWithCancel(r **http.Request) context.CancelFunc {
	ctx, cancel := context.WithCancel((*r).Context())
	*r = (*r).WithContext(ctx)
	return cancel
}

func newSubscriber(r **http.Request, ch chan<- event.Event) event.Subscriber {
	cancel := requestWithCancel(r)

	return event.Subscriber{
		Id:       uuid.New(),
		JoinedAt: time.Now(),
		Address:  (*r).RemoteAddr,
		Chan:     ch,
		Name:     getRequestedSubscriberName(*r),
		Cancel:   cancel,
	}
}
