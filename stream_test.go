package main

import (
	"testing"
)

func TestTransferList(t *testing.T) {

	in1 := make(stream)
	in2 := make(stream)
	inL := []stream{in1, in2}

	suml := func(l []real) real {
		var s real
		for _, r := range l {
			s += r
		}
		return s
	}

	go func() {
		in1 <- 1
		in2 <- 2
	}()

	out := transferList(suml)(inL)
	sum := <-out

	if sum != 3 {
		t.Error("error")
	}

	// the number should be received in any order
	go func() {
		in2 <- 3
		in1 <- 2
	}()

	sum = <-out

	if sum != 5 {
		t.Error("error")
	}

}

func TestSPlitList(t *testing.T) {
	in := make(stream)

	go func() {
		in <- 1
		in <- 2
	}()

	out := splitList(in, 3)
	x0 := <-out[0]
	x1 := <-out[1]
	x2 := <-out[2]

	if x0 != 1 {
		t.Error("Error split")
	}
	if x1 != 1 {
		t.Error("Error split")
	}
	if x2 != 1 {
		t.Error("Error split")
	}
}
