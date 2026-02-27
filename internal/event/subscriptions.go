package event

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type subscriptions struct {
	m  map[string]*subscribers
	mu sync.RWMutex
}

func NewSubscriptions() *subscriptions {
	return &subscriptions{m: make(map[string]*subscribers)}
}

func (s *subscriptions) Load(requestPath string) (Subscribers, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for pattern, subrs := range s.m {
		if ok, _ := match(pattern, requestPath); ok {
			return subrs, true
		}
	}

	return nil, false
}

func (s *subscriptions) Store(pattern string, subr Subscriber) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := match(pattern, ""); err != nil {
		return fmt.Errorf("malformed pattern: %w", err)
	}

	if _, ok := s.m[pattern]; !ok {
		s.m[pattern] = newSubscribers()
	}

	s.m[pattern].Store(subr)
	return nil
}

func (s *subscriptions) Delete(pattern string, subrId uuid.UUID) {
	s.mu.RLock()
	subrs, ok := s.m[pattern]
	s.mu.RUnlock()
	if !ok {
		return
	}

	if subrs.Delete(subrId) > 0 {
		return
	}

	s.mu.Lock()
	if subrs.Len() == 0 {
		delete(s.m, pattern)
	}
	s.mu.Unlock()
}

func (s *subscriptions) Range(f func(pattern string, subrs Subscribers) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for pattern, subrs := range s.m {
		if !f(pattern, subrs) {
			return
		}
	}
}

func match(pattern string, path string) (bool, error) {
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.NewReplacer(
		regexp.QuoteMeta("**"), ".*?",
		regexp.QuoteMeta("*"), "[^/]*",
		regexp.QuoteMeta("?"), "[^/]",
	).Replace(pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("malformed pattern: %w", err)
	}

	return re.MatchString(path), nil
}
