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

	if !s.Contains(value) {
		(*s).mu.Lock()
		hashStr := value.GetHashStr()
		(*s).items[hashStr] = value
		(*s).mu.Unlock()
	}
}

func (s *Set[T]) Remove(value T) {
	if s.Contains(value) {
		(*s).mu.Lock()
		hashStr := value.GetHashStr()
		delete((*s).items, hashStr)
		(*s).mu.Unlock()
	}
}

func (s *Set[T]) Contains(value T) bool {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
	hashStr := value.GetHashStr()
	_, c := (*s).items[hashStr]
	return c
}

func (s *Set[T]) Size() int {
	(*s).mu.Lock()
	defer (*s).mu.Unlock()
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
