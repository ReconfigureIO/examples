package rand

// A Rand is a source of random numbers
type Rand struct {
	seed uint32
}

// New constructs a new Rand given a seed
func New(seed uint32) Rand {
	return Rand{seed: seed}
}

// Uint32s writes a stream of uint32s to the given channel
func (r Rand) Uint32s(output chan<- uint32) {
	seed := r.seed
	for {
		a := seed ^ (seed << 13)
		b := a ^ (a >> 17)
		c := b ^ (b << 5)
		output <- c
		seed = c
	}
}
