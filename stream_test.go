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

	out := splitListN(in, 3)
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

func TestNaturalNumbers(t *testing.T) {
	increase := func(x real) real { return x + 1 }

	nat := recursion(
		func(c stream) stream {
			return prefix(0.0)(transfer(increase)(c))
		})

	zero := <-nat
	if zero != 0 {
		t.Error("Error nat")
	}

	one := <-nat

	if one != 1 {
		t.Error("Error nat")
	}

	for i := 1; i < 100; i++ {
		num := <-nat
		if num != one+real(i) {
			t.Error("Error nat")
		}

	}
}

func TestFactorial1(t *testing.T) {

	nat := natGenerator()
	<-nat
	in := make(stream)
	f := prefix(1.0)(in)
	out := transfer2(
		func(x, y real) real {
			return x * y
		})(nat, f)
	fact, out2 := split(out)
	connect(out2, in)

	fval := []real{1, 2, 6, 24, 120}

	for i := 0; i < len(fval); i++ {
		f := <-fact
		if f != fval[i] {
			t.Errorf("Error fact %f %f", f, fval[i])
		}
	}
}

func TestFactorial2(t *testing.T) {
	nat := natGenerator()
	<-nat

	factorial := recursion(
		func(c stream) stream {
			return transfer2(
				func(x, y real) real {
					return x * y
				})(nat, prefix(1.0)(c))
		})

	fval := []real{1, 2, 6, 24, 120}

	for i := 0; i < len(fval); i++ {
		f := <-factorial
		if f != fval[i] {
			t.Errorf("Error fact %f %f", f, fval[i])
		}
	}
}

func TestNConstantNumbers(t *testing.T) {
	zero := constant(0.0)

	for i := 1; i < 100; i++ {
		num := <-zero
		if num != 0.0 {
			t.Error("Error constant")
		}

	}
}

func TestPairwiseMult(t *testing.T) {

	a := []stream{constant(2.0), constant(5.0)}
	b := []stream{constant(3.0), constant(6.0)}

	n := <-a[0]

	if n != 2.0 {
		t.Error("Error pairwise")
	}

	mult := func(a, b real) real {
		return a * b
	}
	p := pairwise(mult)(a, b)

	n1 := <-p[0]

	if n1 != 6.0 {
		t.Error("Error pairwise")
	}

	done := make(chan bool)
	test := func(ind int, r real) {
		for i := 1; i < 10; i++ {
			num := <-p[ind]
			if num != r {
				t.Errorf("Error pairwise %f, %f", num, r)
			}
		}
		done <- true
	}

	go test(0, 6.0)
	go test(1, 30.0)
	<-done
	<-done
}
