package blockchain

var exists = struct{}{}

type Set struct {
	m map[*Transaction]struct{}
}

func NewSet() *set {
	s := &set{}
	s.m = make(map[*Transaction]struct{})
	return s
}

func (s *set) Add(value *Transaction) {
	s.m[value] = exists
}

func (s *set) Remove(value *Transaction) {
	delete(s.m, value)
}

func (s *set) Contains(value *Transaction) bool {
	_, c := s.m[value]
	return c
}