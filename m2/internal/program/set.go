package program

type Set[K comparable] map[K]struct{}

func NewSet[K comparable]() Set[K] {
	return make(map[K]struct{})
}

// Add a key to the set.
func (s Set[K]) Add(key K) {
	s[key] = struct{}{}
}

// Contains returns true if key is in the set.
func (s Set[K]) Contains(key K) bool {
	if _, exists := s[key]; exists {
		return true
	}
	return false
}
