package twassert

type dev struct {
	*assertions
	number string
}

// Device will allow expecting calls and messages from a particular destination number.
func (a *assertions) Device(number string) Device {
	return &dev{a, number}
}
