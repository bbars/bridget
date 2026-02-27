package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bbars/bridget/internal/event"
)

type ManageList struct {
	Subscriptions Subscriptions
}

type subscriberList struct {
	Pattern     string           `json:"pattern,omitempty"`
	Subscribers []subscriberInfo `json:"subscribers,omitempty"`
}

type subscriberInfo struct {
	Id       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Address  string    `json:"address,omitempty"`
	JoinedAt time.Time `json:"joined_at,omitempty"`
}

func (h ManageList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	var sls []subscriberList
	h.Subscriptions.Range(func(pattern string, subrs event.Subscribers) bool {
		subrSlice := subrs.Load()
		sl := subscriberList{
			Pattern:     pattern,
			Subscribers: make([]subscriberInfo, len(subrSlice)),
		}

		for i, subr := range subrSlice {
			sl.Subscribers[i] = subscriberInfo{
				Id:       subr.Id.String(),
				Name:     subr.Name,
				Address:  subr.Address,
				JoinedAt: subr.JoinedAt,
			}
		}

		sls = append(sls, sl)
		return true
	})

	switch matchContentType(r.Header.Get("Accept"), "application/json", "text/plain") {
	case "application/json":
		h.renderJson(w, sls)
	default:
		h.renderText(w, sls)
	}
}

func (h ManageList) renderJson(w http.ResponseWriter, sls []subscriberList) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(sls); err != nil {
		log.Printf("failed to write response body: %v", err)
	}
}

func (h ManageList) renderText(w http.ResponseWriter, sls []subscriberList) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	n := 0
	now := time.Now()
	for _, sl := range sls {
		for _, subr := range sl.Subscribers {
			n++
			_, err := fmt.Fprintf(w, "%00d.\tid=%s\tjoined_at=%s\tlifetime=%s\taddress=%s\tname=%s\tpattern=%s\n",
				n,
				subr.Id,
				subr.JoinedAt.Format(time.RFC3339),
				now.Sub(subr.JoinedAt).Truncate(time.Second),
				subr.Address,
				subr.Name,
				sl.Pattern,
			)
			if err != nil {
				log.Printf("failed to write response body: %v", err)
				return
			}
		}
	}
}
