package bitset

import (
	"fmt"
	"math/bits"
)

type BitSet struct {
	data []uint64
	size uint64
}

func New(size uint64) *BitSet {
	words := (size + 63) / 64
	return &BitSet{
		data: make([]uint64, words),
		size: size,
	}
}

func (bs *BitSet) Set(pos uint64) {
	if pos >= bs.size {
		return
	}
	word := pos / 64
	bit := pos % 64
	bs.data[word] |= 1 << bit
}

func (bs *BitSet) Get(pos uint64) bool {
	if pos >= bs.size {
		return false
	}
	word := pos / 64
	bit := pos % 64
	return (bs.data[word] & (1 << bit)) != 0
}

func (bs *BitSet) Count() uint {
	count := uint(0)
	for _, word := range bs.data {
		count += uint(bits.OnesCount64(word))
	}
	return count
}

func (bs *BitSet) Size() uint64 {
	return bs.size
}

func (bs *BitSet) Data() []uint64 {
	return bs.data
}

func (bs *BitSet) SetData(data []uint64) error {
	expectedWords := (bs.size + 63) / 64
	if uint64(len(data)) != expectedWords {
		return fmt.Errorf("invalid data length: expected %d words, got %d",
			expectedWords, len(data))
	}
	bs.data = data
	return nil
}

func (bs *BitSet) Clear() {
	for i := range bs.data {
		bs.data[i] = 0
	}
}
