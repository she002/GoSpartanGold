package blockchain

import "sync"

type SetItemInterface interface {
	GetHashStr() string
}

type Set[T SetItemInterface] struct {
	items map[string]T
	mu    sync.Mutex
}

func NewSet[T SetItemInterface]() *Set[T] {
	var newSet Set[T]
	newSet.items = make(map[string]T)
	return &newSet
}

func (s *Set[T]) Add(value T) {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
	if !s.Contains(value) {
		hashStr := value.GetHashStr()
		(*s).items[hashStr] = value
	}
}

func (s *Set[T]) Remove(value T) {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
	if s.Contains(value) {
		hashStr := value.GetHashStr()
		delete((*s).items, hashStr)
	}
}

func (s *Set[T]) Contains(value T) bool {
	hashStr := value.GetHashStr()
	_, c := (*s).items[hashStr]
	return c
}

func (s *Set[T]) Size() int {
	return len((*s).items)
}

func (s *Set[T]) ToArray() []T {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
	var arr []T
	for _, v := range (*s).items {
		arr = append(arr, v)
	}
	return arr
}

func (s *Set[T]) Clear() {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
	(*s).items = make(map[string]T)
}
