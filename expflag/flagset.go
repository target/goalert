package expflag

type FlagSet []Flag

// Has returns true if the given flag name is in the set.
//
// It will panic if the flag name is not known.
func (f FlagSet) Has(flag Flag) bool {
	for _, v := range f {
		if v != flag {
			continue
		}

		return true
	}

	return false
}
