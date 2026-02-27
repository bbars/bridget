package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bbars/bridget/internal/event"
	"github.com/google/uuid"
)

type ManageKick struct {
	Subscriptions Subscriptions
}

func (h ManageKick) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	var kickId uuid.UUID
	if v, err := uuid.Parse(r.PathValue(WildcardId)); err != nil {
		err = fmt.Errorf("malformed subscriber id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		kickId = v
	}

	found := false
	h.Subscriptions.Range(func(pattern string, subrs event.Subscribers) bool {
		for _, subr := range subrs.Load() {
			if subr.Id == kickId {
				log.Printf("kick subscriber id=%s name=%q", subr.Id, subr.Name)
				subr.Cancel()
				found = true
				return false
			}
		}

		return true
	})

	if !found {
		err := fmt.Sprintf("subscriber id=%s not found", kickId)
		http.Error(w, err, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	return
}
