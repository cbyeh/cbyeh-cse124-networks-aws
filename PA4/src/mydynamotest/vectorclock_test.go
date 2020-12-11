package mydynamotest

import (
	"mydynamo"
	"testing"
)

func TestBasicVectorClock(t *testing.T) {
	t.Logf("Starting TestBasicVectorClock")

	//create two vector clocks
	clock1 := mydynamo.NewVectorClock()
	clock2 := mydynamo.NewVectorClock()

	//Test for equality
	if !clock1.Equals(clock2) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
	}

}

// func TestVectorClockEquals(t *testing.T) {

// 	//create two vector clocks
// 	clock1 := mydynamo.NewVectorClock()
// 	clock2 := mydynamo.NewVectorClock()

// 	clock1.Increment("id1")
// 	clock1.Increment("id2")
// 	clock1.Increment("id3")

// 	clock2.Increment("id1")
// 	clock2.Increment("id2")
// 	clock2.Increment("id3")

// 	//Test for equality
// 	if !clock1.Equals(clock2) {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// }

// func TestVectorClockLessThan(t *testing.T) {

// 	//create two vector clocks
// 	clock1 := mydynamo.NewVectorClock()
// 	clock2 := mydynamo.NewVectorClock()

// 	clock1.Increment("id1")
// 	clock1.Increment("id2")
// 	// clock1.Increment("id3")

// 	clock2.Increment("id1")
// 	clock2.Increment("id2")
// 	clock2.Increment("id3")

// 	//Test for equality
// 	if !clock1.LessThan(clock2) {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// }


// func TestVectorClockConcurrent(t *testing.T) {

// 	//create two vector clocks
// 	clock1 := mydynamo.NewVectorClock()
// 	clock2 := mydynamo.NewVectorClock()

// 	clock1.Increment("id1")
// 	clock1.Increment("id1")

// 	clock2.Increment("id1")
// 	clock2.Increment("id2")
// 	clock2.Increment("id2")

// 	//Test for equality
// 	if !clock1.Concurrent(clock2) {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// }


// func TestVectorClockCombine(t *testing.T) {

// 	//create two vector clocks
// 	clock1 := mydynamo.NewVectorClock()
// 	clock2 := mydynamo.NewVectorClock()

// 	clock1.Increment("id1")
// 	clock1.Increment("id2")
// 	clock1.Increment("id3")
// 	clock1.Increment("id4")

// 	clock2.Increment("id5")
// 	clock2.Increment("id2")
// 	clock2.Increment("id2")
// 	clock2.Increment("id2")
// 	clock2.Increment("id6")

// 	arr := clock1.GetArr(clock2)

// 	clock1.Combine(arr)


// 	//Test for equality
// 	if clock1.PairMap["id1"] != 1 {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// 	if clock1.PairMap["id2"] != 3 {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// 	if clock1.PairMap["id3"] != 1 {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// 	if clock1.PairMap["id4"] != 1 {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// 	if clock1.PairMap["id5"] != 1 {
// 		t.Fail()
// 		t.Logf("Vector Clocks were not equal")
// 	}

// }