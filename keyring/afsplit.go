package keyring

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"fmt"

	"github.com/pkg/errors"
)

func xor(a, b []byte) []byte {
	if len(a) != len(b) {
		// not handling different lengths
		panic(fmt.Sprintf("length mismatch, %d != %d", len(a), len(b)))
	}
	out := make([]byte, len(a))
	for i := range out {
		out[i] = a[i] ^ b[i]
	}
	return out
}

func diffuse(block []byte, digest crypto.Hash) []byte {
	fullBlocks := len(block) / digest.Size()
	lastBlockSize := len(block) % digest.Size()

	out := make([]byte, len(block))
	copy(out, block)
	hash := digest.New()
	for i := 0; i < fullBlocks; i++ {
		hash.Reset()
		hash.Write(out)
		copy(out[i*digest.Size():], hash.Sum(nil))
	}

	if lastBlockSize == 0 {
		return out
	}

	hash.Reset()
	hash.Write(out)
	copy(out[fullBlocks*digest.Size():], hash.Sum(nil))
	return out[:len(block)]
}

func AFSplit(data []byte, stripes int, digest crypto.Hash) ([]byte, error) {
	if stripes < 2 {
		return nil, errors.New("number of stripes must be >= 2")
	}
	var randData [][]byte
	state := make([]byte, len(data))
	for i := 1; i < stripes; i++ {
		buf := make([]byte, len(data))
		rand.Read(buf)
		randData = append(randData, buf)
		state = xor(state, buf)
		state = diffuse(state, digest)
	}

	ciphertext := xor(state, data)
	out := bytes.NewBuffer(make([]byte, 0, len(data)*(stripes+1)))
	out.Write(ciphertext)

	state = diffuse(ciphertext, digest)
	for _, d := range randData {
		out.Write(xor(state, d))
	}

	return out.Bytes(), nil
}

func AFMerge(data []byte, stripes int, digest crypto.Hash) ([]byte, error) {
	if stripes < 2 {
		return nil, errors.New("number of stripes must be >= 2")
	}
	if len(data)%stripes != 0 {
		return nil, errors.New("invalid length for number of stripes")
	}
	r := bytes.NewReader(data)
	state := make([]byte, len(data)/stripes)
	ciphertext := make([]byte, len(state))
	currentChunk := make([]byte, len(state))

	r.Read(ciphertext)
	blockKey := diffuse(ciphertext, digest)

	for i := 1; i < stripes; i++ {
		r.Read(currentChunk)
		currentChunk = xor(blockKey, currentChunk)

		state = xor(state, currentChunk)
		state = diffuse(state, digest)
	}

	return xor(ciphertext, state), nil
}
