package swogrp

import (
	"sync"

	"github.com/google/uuid"
)

type Set struct {
	m  map[uuid.UUID]struct{}
	mx sync.Mutex
}

func NewSet() *Set {
	return &Set{
		m: make(map[uuid.UUID]struct{}),
	}
}

func (s *Set) Add(id uuid.UUID) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.m[id] = struct{}{}
}

func (s *Set) Has(id uuid.UUID) bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, ok := s.m[id]
	return ok
}

func (s *Set) List() []uuid.UUID {
	s.mx.Lock()
	defer s.mx.Unlock()

	ids := make([]uuid.UUID, 0, len(s.m))
	for id := range s.m {
		ids = append(ids, id)
	}

	return ids
}
