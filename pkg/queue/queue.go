package queue

import (
	"container/list"
	"sync"
)

type StatusQueue struct {
	sync.RWMutex
	items                *list.List
	completedOptionCount uint
	currentCount         uint
	fireFunc             func()
}

func New() *StatusQueue {
	return &StatusQueue{items: list.New()}
}

func (s *StatusQueue) RegistFire(completedOption uint, f func()) {
	s.completedOptionCount = completedOption
	s.fireFunc = f
}

func (s *StatusQueue) Enqueue(item interface{}) {
	s.Lock()
	defer s.Unlock()

	s.items.PushBack(item)

	if s.fireFunc != nil {
		s.currentCount++

		if s.currentCount >= s.completedOptionCount {
			go s.fireFunc()
			s.removeAll()
		}
	}
}

func (s *StatusQueue) Dequeue() interface{} {
	return nil
}

func (s *StatusQueue) removeAll() {
	var next *list.Element
	for e := s.items.Front(); e != nil; e = next {
		next = e.Next()
		s.items.Remove(e)
	}

	s.currentCount = 0
}

func (s *StatusQueue) RemoveAll() {
	s.Lock()
	defer s.Unlock()

	if s.items.Len() == 0 {
		return
	}

	s.removeAll()

}
