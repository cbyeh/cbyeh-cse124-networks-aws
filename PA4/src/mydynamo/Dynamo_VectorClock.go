package mydynamo

type VectorClock struct {
	pairArray []Pair
}

type Pair struct {
	clockNumber int
	hostID      string
}

// Creates a new VectorClock
func NewVectorClock() VectorClock {
	newPairArray := make([]Pair, 5)
	for i := 0; i < 5; i++ {
		newPairArray[i] = Pair{0, ""}
	}
	var newVectorClock VectorClock
	newVectorClock.pairArray = newPairArray
	return newVectorClock
}

// Returns true if the other VectorClock is causally descended from this one
func (s VectorClock) LessThan(otherClock VectorClock) bool {
	// V(a) < V(b) when a_k <= b_k for all k and V(a) != V(b)
	for i := 0; i < 5; i++ {
		if s.pairArray[i].clockNumber > otherClock.pairArray[i].clockNumber {
			return false
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
	lessThan := false
	greaterThan := false
	for i := 0; i < 5; i++ {
		if s.pairArray[i].clockNumber < otherClock.pairArray[i].clockNumber {
			lessThan = true
		} else if s.pairArray[i].clockNumber > otherClock.pairArray[i].clockNumber {
			greaterThan = true
		}
		if lessThan && greaterThan {
			return true
		}
	}
	return false
	// Note: Could call Less than (a, b) and (b, a) and check if both return false
}

// Increments this VectorClock at the element associated with nodeId
func (s *VectorClock) Increment(nodeId string) {
	for _, pair := range s.pairArray {
		if pair.hostID == nodeId {
			pair.clockNumber++
		}
	}
}

// Changes this VectorClock to be causally descended from all VectorClocks in clocks
func (s *VectorClock) Combine(clocks []VectorClock) {
	// Return a vector clock >= all clocks, including s
	var newVectorClock VectorClock
	newVectorClock = *s
	for _, c := range clocks {
		pairArray := c.pairArray
		for j, p := range pairArray {
			if p.clockNumber > newVectorClock.pairArray[j].clockNumber {
				newVectorClock.pairArray[j].clockNumber = p.clockNumber
			}
		}
	}
	*s = newVectorClock
}

// Tests if two VectorClocks are equal
func (s VectorClock) Equals(otherClock VectorClock) bool {
	// V(a) = V(b) when a_k = b_k for all k
	for i := 0; i < 5; i++ {
		if s.pairArray[i].clockNumber != otherClock.pairArray[i].clockNumber {
			return false
		}
	}
	return true
}
