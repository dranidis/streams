package main

import (
	"fmt"
)

type freal func(float64) float64
type freal2 func(float64, float64) float64
type rchan chan float64
type fchan func(rchan) rchan
type fchan2 func(rchan, rchan) rchan

func transfer(f freal) fchan {
	return func(in rchan) rchan {
		out := make(rchan)
		go func() {
			for {
				x := <-in
				out <- f(x)
			}
		}()
		return out
	}
}

func transfer2(f freal2) fchan2 {
	return func(in1, in2 rchan) rchan {
		out := make(rchan)
		go func() {
			for {
				var x1, x2 float64
				select {
				case x1 = <-in1:
					x2 = <-in2
				case x2 = <-in2:
					x1 = <-in1
				}
				out <- f(x1, x2)
			}
		}()
		return out
	}
}

func prefix(pre float64) fchan {
	return func(in rchan) rchan {
		out := make(rchan)
		go func() {
			out <- pre
			for {
				x := <-in
				out <- x
			}
		}()
		return out
	}
}

func split(in rchan) (rchan, rchan) {
	out1 := make(rchan)
	out2 := make(rchan)
	go func() {
		for {
			x := <-in
			select {
			case out1 <- x:
				out2 <- x
			case out1 <- x:
				out2 <- x
			}
		}
	}()
	return out1, out2

}

func connect(in rchan, out rchan) {
	go func() {
		for {
			x := <-in
			out <- x
		}
	}()
}

func recursion(f fchan) rchan {
	in := make(rchan)
	out := f(in)
	outS, feedback := split(out)
	connect(feedback, in)
	return outS
}

func main() {
	in := make(rchan)
	increase := func(x float64) float64 { return x + 1 }
	add := func(x, y float64) float64 { return x + y }

	out := prefix(3.0)(transfer(increase)(in))

	go func() {
		in <- 1.0
	}()

	fmt.Println(<-out)
	fmt.Println(<-out)

	nat := recursion(
		func(c rchan) rchan {
			return prefix(0.0)(transfer(increase)(c))
		})

	in1 := make(rchan)
	in2 := make(rchan)
	go func() {
		for {
			in1 <- 1.0
		}
	}()
	go func() {
		for {
			in2 <- 2.0
		}
	}()

	out2 := transfer2(add)(in1, nat)

	fmt.Println(<-nat)
	fmt.Println(<-nat)
	fmt.Println(<-nat)

	fmt.Println(<-out2)
	fmt.Println(<-out2)
	fmt.Println(<-out2)

}
