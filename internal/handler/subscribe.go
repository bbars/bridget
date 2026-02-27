package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bbars/bridget/internal/event"
	"github.com/bbars/bridget/internal/forwarder"
)

type Subscribe struct {
	Subscriptions Subscriptions
}

func (h Subscribe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	if r.Method != http.MethodGet {
		http.Error(w, "only method GET is supported", http.StatusMethodNotAllowed)
		return
	}

	var accept string
	if r.URL.Query().Has("accept") {
		accept = r.URL.Query().Get("accept")
	} else {
		accept = r.Header.Get("Accept")
	}

	ch := make(chan event.Event)
	defer close(ch)

	subr := newSubscriber(&r, ch)
	defer subr.Cancel()

	pathPattern := r.PathValue(WildcardPath)
	if err := h.Subscriptions.Store(pathPattern, subr); err != nil {
		err = fmt.Errorf("register pattern: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer h.Subscriptions.Delete(pathPattern, subr.Id)

	log.Printf("new subscriber id=%s (name=%q) for %q\n", subr.Id, subr.Name, pathPattern)
	defer log.Printf("subscriber id=%s (name=%q) for %q is disconnected\n", subr.Id, subr.Name, pathPattern)

	switch {
	case accept == forwarder.SseMimeType:
		if err := forwarder.ForwardSse(subr, ch, w, r); err != nil {
			log.Printf("failed to forward SSE: %v\n", err)
		}
	default:
		if err := forwarder.ForwardMultipartRaw(subr, ch, w, r); err != nil {
			log.Printf("failed to forward multipart/raw: %v\n", err)
		}
	}
}
