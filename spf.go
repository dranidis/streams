package streams

// NatGenerator generates a stream of numbers starting from n and incremented by 1.0
func NatGenerator(n Real) Stream {
	succ := func(x Real) Real { return x + 1 }
	return recursion(
		func(c Stream) Stream {
			return Prefix(n)(Transfer(succ)(c))
		})
}
