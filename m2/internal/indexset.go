package internal

type indexSet map[int]struct{}

func newIndexSet() indexSet {
	return make(map[int]struct{})
}

// Add an index to the set.
func (s indexSet) Add(index int) {
	s[index] = struct{}{}
}

func (s indexSet) Contains(index int) bool {
	if _, exists := s[index]; exists {
		return true
	}
	return false
}
