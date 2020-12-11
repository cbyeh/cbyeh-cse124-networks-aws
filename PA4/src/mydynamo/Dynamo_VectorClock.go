package mydynamo
// import (
// 	"log"
// )
type VectorClock struct {
	//todo
	PairMap		map[string]Pair

}

type Pair struct {
	nodeId		string
	ClockNumber	int
}

//Creates a new VectorClock
func NewVectorClock() VectorClock {
	newPairMap := make(map[string]Pair)

	newVectorClock := VectorClock{newPairMap}

	return newVectorClock

}

//Returns true if the other VectorClock is causally descended from this one
func (s VectorClock) LessThan(otherClock VectorClock) bool {

	for nodeId, currPair := range s.PairMap {
		otherPair, ok := otherClock.PairMap[nodeId]
		if ok {
			if otherPair.ClockNumber < currPair.ClockNumber {
				return false
			}
		} else {
			return false
		}
	}

	// for nodeId, otherPair := range otherClock.PairMap {
	// 	currPair, ok := s.PairMap[nodeId]
	// 	if ok {
	// 		if otherPair.ClockNumber < currPair.ClockNumber {
	// 			return false
	// 		}
	// 	}
	// }

	if s.Equals(otherClock) {
		return false
	}

	return true



}

//Returns true if neither VectorClock is causally descended from the other
func (s VectorClock) Concurrent(otherClock VectorClock) bool {

	if (!s.LessThan(otherClock)) && (!otherClock.LessThan(s)) {
		return true
	}
	return false
}

//Increments this VectorClock at the element associated with nodeId
func (s *VectorClock) Increment(nodeId string) {

	pair, ok := s.PairMap[nodeId]
	if ok {
		newPair := Pair{nodeId, pair.ClockNumber + 1}
		s.PairMap[nodeId] = newPair
	} else {
		newPair := Pair{nodeId, 1}
		s.PairMap[nodeId] = newPair
	}
	
	
}

//Changes this VectorClock to be causally descended from all VectorClocks in clocks
func (s *VectorClock) Combine(clocks []VectorClock) {

	maxMap := make(map[string]Pair)
	
	for nodeId, pair := range s.PairMap {
		maxMap[nodeId] = Pair{pair.nodeId, pair.ClockNumber}
	}

	for _, c := range clocks {
		for nodeId, currPair := range c.PairMap {
			maxPair, ok := maxMap[nodeId]
			if (ok && currPair.ClockNumber > maxPair.ClockNumber) || (!ok) {
				newPair := Pair{nodeId, currPair.ClockNumber}
				maxMap[nodeId] = newPair
			}
		}
	}

	var newVectorClock VectorClock
	newVectorClock.PairMap = maxMap

	*s = newVectorClock

}

//Tests if two VectorClocks are equal
func (s *VectorClock) Equals(otherClock VectorClock) bool {
	for nodeId, currPair := range s.PairMap {
		otherPair, ok := otherClock.PairMap[nodeId]
		if ok {
			if currPair.ClockNumber  != otherPair.ClockNumber {
				return false
			}
		} else {
			return false
		}
	}

	for nodeId, otherPair := range otherClock.PairMap {
		currPair, ok := s.PairMap[nodeId]
		if ok {
			if otherPair.ClockNumber != currPair.ClockNumber {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (s *VectorClock) GetArr(c VectorClock) []VectorClock {
	arr := []VectorClock{c}
	return arr
}
