package hasher

import (
	"github.com/spaolacci/murmur3"
)

type Hasher interface {
	Hashes(data []byte, k, m uint64) []uint64
}

type MurmurHasher struct{}

func New() Hasher {
	return &MurmurHasher{}
}

func (mh *MurmurHasher) Hashes(data []byte, k, m uint64) []uint64 {
	h1, h2 := murmur3.Sum128(data)
	hashes := make([]uint64, k)

	for i := uint64(0); i < k; i++ {
		hashes[i] = (h1 + i*h2) % m
	}

	return hashes
}
