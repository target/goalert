package keyring

type KeySlot struct {
	KeyID   string
	SlotID  int
	Version int

	Salt       []byte
	Iterations int

	Stripes int

	Material []byte
}

func (k KeySlot) Decrypt() []byte {
	panic("golang.org/x/crypto/pbkdf2")
}
