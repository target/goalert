package keyring

import (
	"bytes"
	"crypto"
	cRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
)

var pseudoRand *rand.Rand

func init() {
	b := make([]byte, 8)

	_, err := cRand.Read(b)
	if err != nil {
		// lack of entropy
		panic(err)
	}

	pseudoRand = rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(b))))
}

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

func AFSplit(data []byte, stripes int, digest crypto.Hash) []byte {
	out := bytes.NewBuffer(make([]byte, 0, len(data)*stripes))
	state := make([]byte, len(data))
	randBuf := make([]byte, len(data))
	for i := 1; i < stripes; i++ {
		pseudoRand.Read(randBuf)
		out.Write(randBuf)
		state = xor(state, randBuf)
		state = diffuse(state, digest)
	}

	out.Write(xor(state, data))
	return out.Bytes()
}

func AFMerge(data []byte, stripes int, digest crypto.Hash) []byte {
	r := bytes.NewReader(data)
	state := make([]byte, len(data)/stripes)
	currentChunk := make([]byte, len(state))
	for i := 1; i < stripes; i++ {
		r.Read(currentChunk)
		state = xor(currentChunk, state)
		state = diffuse(state, digest)
	}

	n, _ := r.Read(currentChunk)
	return xor(currentChunk[:n], state)
}
