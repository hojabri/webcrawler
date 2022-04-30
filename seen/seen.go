package seen

import (
	"sync"
)

type Seen struct {
	List sync.Map
}

func NewSeenList() Seen {
	return Seen{}
}

func (s *Seen) Store(item any, value any) {
	s.List.Store(item, value)
}

func (s *Seen) Delete(item any) {
	s.List.Delete(item)
}

func (s *Seen) Exist(item any) bool {
	_, ok := s.List.Load(item)
	return ok
}

func (s *Seen) Map() map[any]any {
	m := map[any]any{}
	s.List.Range(func(key, value interface{}) bool {
		m[key] = value
		return true
	})
	return m
}
