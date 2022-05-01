package visited

import (
	"strings"
	"sync"
)

type Visited struct {
	List sync.Map
}

func NewVisitedList() Visited {
	return Visited{}
}

func pureURL(item string) string {
	item = strings.TrimSuffix(item, "/")        // remove /
	item = strings.TrimPrefix(item, "https://") // remove https://
	item = strings.TrimPrefix(item, "http://")  // remove http://
	return item
}

func (s *Visited) Store(item string, value any) {
	item = pureURL(item)
	s.List.Store(item, value)
}

func (s *Visited) Delete(item string) {
	item = pureURL(item)
	s.List.Delete(item)
}

func (s *Visited) Exist(item string) bool {
	item = pureURL(item)
	_, ok := s.List.Load(item)
	return ok
}

func (s *Visited) Map() map[any]any {
	m := map[any]any{}
	s.List.Range(func(key, value interface{}) bool {
		m[key] = value
		return true
	})
	return m
}
