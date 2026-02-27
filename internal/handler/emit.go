package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bbars/bridget/internal/event"
)

type Emit struct {
	Subscriptions Subscriptions
}

func (h Emit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	requestPath := r.PathValue(WildcardPath)
	if err := h.emitEvent(requestPath, w, r); err != nil {
		log.Printf("failed to receive event %q: %v\n", requestPath, err)
	}
}

func (h Emit) emitEvent(requestPath string, w http.ResponseWriter, r *http.Request) error {
	var subs event.Subscribers
	if v, ok := h.Subscriptions.Load(requestPath); !ok {
		err := fmt.Errorf("no subscribers found for %s", requestPath)
		http.Error(w, err.Error(), http.StatusNotFound)
		return err
	} else {
		subs = v
	}

	forwardResults := make(chan event.ForwardResult)
	defer func() {
		close(forwardResults)
	}()

	evt := event.Event{
		Data: event.EventData{
			ClientAddr: r.RemoteAddr,
			Method:     r.Method,
			Path:       requestPath,
			Query:      r.URL.RawQuery,
			Header:     r.Header.Clone(),
		},
		ForwardResults: forwardResults,
	}
	if body, err := io.ReadAll(r.Body); err != nil {
		err = fmt.Errorf("read body: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	} else {
		evt.Data.Body = body
	}

	w.WriteHeader(http.StatusOK)

	log.Printf("forward event for %q", requestPath)
	cnt := 0
	for _, sub := range subs.Load() {
		select {
		case <-r.Context().Done():
			return r.Context().Err()
		case sub.Chan <- evt:
			cnt++
		}
	}

	for i := 0; i < cnt; i++ {
		select {
		case <-r.Context().Done():
			return r.Context().Err()
		case fwRes := <-forwardResults:
			var statusLine string
			if fwRes.Err != nil {
				statusLine = fwRes.Err.Error()
			} else {
				statusLine = "ok"
			}

			if _, err := fmt.Fprintln(w, fwRes.SubscriberId.String()+": "+statusLine); err != nil {
				log.Printf("failed to write response body: %v", err)
			}
		}
	}

	return nil
}
