package blockchain

type SetItemInterface interface {
	GetHashStr() string
}

type Set[T SetItemInterface] struct {
	items map[string]T
}

func NewSet[T SetItemInterface]() *Set[T] {
	var newSet Set[T]
	newSet.items = make(map[string]T)
	return &newSet
}

func (s *Set[T]) Add(value T) {
	if !s.Contains(value) {
		hashStr := value.GetHashStr()
		s.items[hashStr] = value
	}
}

func (s *Set[T]) Remove(value T) {
	if s.Contains(value) {
		hashStr := value.GetHashStr()
		delete(s.items, hashStr)
	}
}

func (s *Set[T]) Contains(value T) bool {
	hashStr := value.GetHashStr()
	_, c := s.items[hashStr]
	return c
}

func (s *Set[T]) Size() int {
	return len(s.items)
}

func (s *Set[T]) ToArray() []T {
	var arr []T
	for _, v := range s.items {
		arr = append(arr, v)
	}
	return arr
}
