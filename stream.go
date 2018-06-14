package streams

import (
	"fmt"
	"reflect"
)

// Real is the main type used in all Stream processing functions
// changing this will globally change the type of data.
type Real float32

// Stream is the main type for stream processing functions.
type Stream chan Real

// Constant creates a Stream consisting of the same number.
func Constant(n Real) Stream {
	out := make(Stream)
	go func() {
		for {
			out <- n
		}
	}()
	return out
}

// Transfer lifts a unary real function to a unary stream function
func Transfer(f func(Real) Real) func(Stream) Stream {
	return func(in Stream) Stream {
		out := make(Stream)
		go func() {
			for {
				out <- f(<-in)
			}
		}()
		return out
	}
}

// Transfer2 lifts a binary real function to a binary stream function
func Transfer2(f func(Real, Real) Real) func(Stream, Stream) Stream {
	return func(in1, in2 Stream) Stream {
		out := make(Stream)
		go func() {
			for {
				var x1, x2 Real
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

// TransferList lifts a list processing function to a list stream function
func TransferList(f func([]Real) Real) func([]Stream) Stream {
	return func(inl []Stream) Stream {
		out := make(Stream)
		go func() {
			for {
				out <- f(readChannelList(inl))
			}
		}()
		return out
	}
}

// Prefix adds an element to the beginning of a stream
func Prefix(pre Real) func(Stream) Stream {
	return func(in Stream) Stream {
		out := make(Stream)
		go func() {
			out <- pre
			for {
				out <- <-in
			}
		}()
		return out
	}
}

func pairwise(f func(Real, Real) Real) func([]Stream, []Stream) []Stream {
	return func(a, b []Stream) []Stream {
		l := len(a)
		out := make([]Stream, l)
		for i := 0; i < l; i++ {
			out[i] = Transfer2(f)(a[i], b[i])
		}
		return out
	}
}

// read a single value from each channel from a list of channels
// in any order.
func readChannelList(inl []Stream) []Real {
	l := len(inl)
	vs := make([]Real, l)
	cases := make([]reflect.SelectCase, l)
	for i, ch := range inl {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, _ := reflect.Select(cases)
		// set the value to nil so as not to read again from this channel
		cases[chosen].Chan = reflect.ValueOf(nil)
		remaining--
		vs[chosen] = value.Interface().(Real)
	}
	return vs
}

func split(in Stream) (Stream, Stream) {
	out1 := make(Stream)
	out2 := make(Stream)
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

func splitList(in []Stream) ([]Stream, []Stream) {
	l := len(in)
	out1 := make([]Stream, l)
	out2 := make([]Stream, l)
	for i := 1; i < l; i++ {
		out1[i], out2[i] = split(in[i])
	}
	return out1, out2
}

func splitListN(in Stream, l int) []Stream {
	out := make([]Stream, l)
	for i := 0; i < l; i++ {
		out[i] = make(Stream)
	}
	go func() {
		for {
			x := <-in
			for i := 0; i < l; i++ {
				out[i] <- x
			}
		}
	}()
	return out
}

func connect(in Stream, out Stream) {
	go func() {
		for {
			out <- <-in
		}
	}()
}

func connectList(in []Stream, out []Stream) {
	l := len(in)
	for i := 1; i < l; i++ {
		connect(in[i], out[i])
	}
}

func recursionList(f func(Stream) Stream, l int) []Stream {
	out := make([]Stream, l)
	for i := 1; i < l; i++ {
		out[i] = recursion(f)
	}
	return out
}

func recursion(f func(Stream) Stream) Stream {
	in := make(Stream)
	out := f(in)
	outS, feedback := split(out)
	connect(feedback, in)
	return outS
}

func main() {
	in := make(Stream)
	increase := func(x Real) Real { return x + 1 }
	add := func(x, y Real) Real { return x + y }

	out := Prefix(3.0)(Transfer(increase)(in))

	go func() {
		in <- 1.0
	}()

	fmt.Println(<-out)
	fmt.Println(<-out)

	nat := recursion(
		func(c Stream) Stream {
			return Prefix(0.0)(Transfer(increase)(c))
		})

	in1 := make(Stream)
	in2 := make(Stream)
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

	out2 := Transfer2(add)(in1, nat)

	fmt.Println(<-nat)
	fmt.Println(<-nat)
	fmt.Println(<-nat)

	fmt.Println(<-out2)
	fmt.Println(<-out2)
	fmt.Println(<-out2)

	suml := func(l []Real) Real {
		var s Real
		for _, r := range l {
			s += r
		}
		return s
	}

	chanList := []Stream{in1, in2}
	out3 := TransferList(suml)(chanList)
	fmt.Println(<-out3)
	fmt.Println(<-out3)
	fmt.Println(<-out3)
}
