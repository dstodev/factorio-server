package program

type IndexSet map[int]struct{}

func NewIndexSet() IndexSet {
	return make(map[int]struct{})
}

// Add an index to the set.
func (s IndexSet) Add(index int) {
	s[index] = struct{}{}
}

// Contains returns true if the index is in the set.
func (s IndexSet) Contains(index int) bool {
	if _, exists := s[index]; exists {
		return true
	}
	return false
}
