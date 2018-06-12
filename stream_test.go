package main

import (
	"fmt"
	"testing"
)

func TestTransferList(t *testing.T) {

	in1 := make(rchan)
	in2 := make(rchan)
	inL := []rchan{in1, in2}

	suml := func(l []float64) float64 {
		var s float64
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

	fmt.Println(sum)

	if sum != 3 {
		t.Error("error")
	}

	go func() {
		in2 <- 3
		in1 <- 2
	}()

	sum = <-out

	if sum != 5 {
		t.Error("error")
	}

}
