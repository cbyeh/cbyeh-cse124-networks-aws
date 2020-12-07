package mydynamo

type VectorClock struct {
	PairMap map[string]int
}

// Creates a new VectorClock
func NewVectorClock() VectorClock {
	newVectorClock := VectorClock{
		PairMap: make(map[string]int),
	}
	return newVectorClock
}

// Returns true if the other VectorClock is causally descended from this one
func (s VectorClock) LessThan(otherClock VectorClock) bool {
	// V(a) < V(b) when a_k <= b_k for all k and V(a) != V(b)
	for k, v := range s.PairMap {
		otherClockNum, ok := otherClock.PairMap[k]
		if ok {
			if v > otherClockNum {
				return false
			}
		} else {
			return false
		}
	}
	for k, v := range otherClock.PairMap {
		currClockNum, ok := s.PairMap[k]
		if ok {
			if v < currClockNum {
				return false
			}
		}
	}
	if s.Equals(otherClock) {
		return false
	}
	return true
}

// Returns true if neither VectorClock is causally descended from the other
func (s VectorClock) Concurrent(otherClock VectorClock) bool {
	// a || b if a_i < b_i and a_j > b_j for some i, j
	if (!s.LessThan(otherClock)) && (!otherClock.LessThan(s)) {
		return true
	}
	return false
}

// Increments this VectorClock at the element associated with nodeId
func (s *VectorClock) Increment(nodeId string) {
	_, ok := s.PairMap[nodeId]
	if ok {
		s.PairMap[nodeId]++
	} else {
		s.PairMap[nodeId] = 1
	}
}

// Changes this VectorClock to be causally descended from all VectorClocks in clocks
func (s *VectorClock) Combine(clocks []VectorClock) {
	// Return a vector clock >= all clocks, including s
	maxMap := make(map[string]int)
	for k, v := range s.PairMap {
		maxMap[k] = v
	}
	for _, c := range clocks {
		cMap := c.PairMap
		for k, v := range cMap {
			maxClockNum, ok := maxMap[k]
			if (ok && v > maxClockNum) || (!ok) {
				maxMap[k] = v
			}
		}
	}
	var newVectorClock VectorClock
	newVectorClock.PairMap = maxMap
	*s = newVectorClock
}

// Tests if two VectorClocks are equal
func (s VectorClock) Equals(otherClock VectorClock) bool {
	// V(a) = V(b) when a_k == b_k for all k
	for k, v := range s.PairMap {
		otherClockNum, ok := otherClock.PairMap[k]
		if ok {
			if v != otherClockNum {
				return false
			}
		} else {
			return false
		}
	}
	for k, v := range otherClock.PairMap {
		currClockNum, ok := s.PairMap[k]
		if ok {
			if v != currClockNum {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
