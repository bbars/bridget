package event

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type Subscribers interface {
	Store(subr Subscriber)
	Load() []Subscriber
}

type subscribers struct {
	items atomic.Value
	mu    sync.Mutex
}

func newSubscribers() *subscribers {
	items := atomic.Value{}
	items.Store(make([]Subscriber, 0))
	return &subscribers{
		items: items,
	}
}

func (s *subscribers) Store(subr Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.items.Load().([]Subscriber)
	for i := range items {
		if items[i].Id == subr.Id {
			items[i] = subr // replace
			s.items.Store(items)
			return
		}
	}

	items = append(items, subr)
	s.items.Store(items)
}

func (s *subscribers) Load() []Subscriber {
	return s.items.Load().([]Subscriber)
}

func (s *subscribers) Len() int {
	return len(s.items.Load().([]Subscriber))
}

func (s *subscribers) Delete(subrId uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.items.Load().([]Subscriber)
	for i := range items {
		if items[i].Id == subrId {
			items = append(items[:i], items[i+1:]...)
			s.items.Store(items)
			return len(items)
		}
	}

	return len(items)
}
