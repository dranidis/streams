package main

import (
	"fmt"
	"reflect"
)

// this is the main type used in all stream processing functions
// changing this will globally change the type of data
type real float32
type stream chan real

func constant(n real) stream {
	out := make(stream)
	go func() {
		for {
			out <- n
		}
	}()
	return out
}

func natGenerator() stream {
	increase := func(x real) real { return x + 1 }

	return recursion(
		func(c stream) stream {
			return prefix(0.0)(transfer(increase)(c))
		})
}

func pairwise(f func(real, real) real) func([]stream, []stream) []stream {
	return func(a, b []stream) []stream {
		l := len(a)
		out := make([]stream, l)
		for i := 0; i < l; i++ {
			out[i] = transfer2(f)(a[i], b[i])
		}
		return out
	}
}

func transfer(f func(real) real) func(stream) stream {
	return func(in stream) stream {
		out := make(stream)
		go func() {
			for {
				out <- f(<-in)
			}
		}()
		return out
	}
}

func transfer2(f func(real, real) real) func(stream, stream) stream {
	return func(in1, in2 stream) stream {
		out := make(stream)
		go func() {
			for {
				var x1, x2 real
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

// read a single value from each channel from a list of channels
// in any order.
func readChannelList(inl []stream) []real {
	l := len(inl)
	vs := make([]real, l)
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
		vs[chosen] = value.Interface().(real)
	}
	return vs
}

func transferList(f func([]real) real) func([]stream) stream {
	return func(inl []stream) stream {
		out := make(stream)

		go func() {
			for {
				out <- f(readChannelList(inl))
			}
		}()
		return out
	}
}

func prefix(pre real) func(stream) stream {
	return func(in stream) stream {
		out := make(stream)
		go func() {
			out <- pre
			for {
				out <- <-in
			}
		}()
		return out
	}
}

func split(in stream) (stream, stream) {
	out1 := make(stream)
	out2 := make(stream)
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

func splitList(in []stream) ([]stream, []stream) {
	l := len(in)
	out1 := make([]stream, l)
	out2 := make([]stream, l)
	for i := 1; i < l; i++ {
		out1[i], out2[i] = split(in[i])
	}
	return out1, out2
}

func splitListN(in stream, l int) []stream {
	out := make([]stream, l)
	for i := 0; i < l; i++ {
		out[i] = make(stream)
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

func connect(in stream, out stream) {
	go func() {
		for {
			out <- <-in
		}
	}()
}

func connectList(in []stream, out []stream) {
	l := len(in)
	for i := 1; i < l; i++ {
		connect(in[i], out[i])
	}
}

func recursionList(f func(stream) stream, l int) []stream {
	out := make([]stream, l)
	for i := 1; i < l; i++ {
		out[i] = recursion(f)
	}
	return out
}

func recursion(f func(stream) stream) stream {
	in := make(stream)
	out := f(in)
	outS, feedback := split(out)
	connect(feedback, in)
	return outS
}

func main() {
	in := make(stream)
	increase := func(x real) real { return x + 1 }
	add := func(x, y real) real { return x + y }

	out := prefix(3.0)(transfer(increase)(in))

	go func() {
		in <- 1.0
	}()

	fmt.Println(<-out)
	fmt.Println(<-out)

	nat := recursion(
		func(c stream) stream {
			return prefix(0.0)(transfer(increase)(c))
		})

	in1 := make(stream)
	in2 := make(stream)
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

	suml := func(l []real) real {
		var s real
		for _, r := range l {
			s += r
		}
		return s
	}

	chanList := []stream{in1, in2}
	out3 := transferList(suml)(chanList)
	fmt.Println(<-out3)
	fmt.Println(<-out3)
	fmt.Println(<-out3)
}
