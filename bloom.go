/*
Package bitbloom provides a high-performance, thread-safe Bloom filter implementation in Go.

A Bloom filter is a probabilistic data structure used to test whether an element is a member of a set.
False positives are possible, but false negatives are not. This means that it can tell you with certainty
that an element is not in the set, but it may incorrectly report that an element is present.

Features:
  - Optimal bit array size and hash function count calculation
  - Thread-safe for concurrent Add and Test operations
  - Accurate fill ratio estimation and tracking
  - Memory-efficient storage using a compact bitset
  - Binary serialization for persistence or transfer

Usage:

	package main

	import (
		"fmt"
		"log"

		"github.com/umang-sinha/bitbloom"
	)

	func main() {
		bf, err := bitbloom.New(1000000, 0.01) // 1 million items, 1% false positive rate
		if err != nil {
			log.Fatal(err)
		}

		bf.Add([]byte("golang"))
		if bf.Test([]byte("golang")) {
			fmt.Println("Item possibly present")
		}
	}
*/
package bitbloom

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/umang-sinha/bitbloom/internal/bitset"
	"github.com/umang-sinha/bitbloom/internal/hasher"
)

// OptimalM calculates the optimal size of the bit array (m) given the expected number
// of items (n) and the desired false positive probability (p).
// This formula is based on the standard Bloom filter size formula:
//
//	m = -(n * ln(p)) / (ln(2)^2)
func OptimalM(n uint64, p float64) uint64 {
	return uint64(math.Ceil(-float64(n) * math.Log(p) / (math.Log(2) * math.Log(2))))
}

// OptimalK calculates the optimal number of hash functions (k) given the bit array size (m)
// and the expected number of inserted items (n).
// This is based on the formula:
//
//	k = (m / n) * ln(2)
func OptimalK(m, n uint64) uint64 {
	return uint64(math.Ceil(float64(m) / float64(n) * math.Log(2)))
}

// BloomFilter represents a Bloom filter instance.
// It is safe for concurrent use by multiple goroutines.
type BloomFilter struct {
	bitset *bitset.BitSet
	hasher hasher.Hasher
	mutex  sync.RWMutex
	m      uint64
	k      uint64
	count  uint64
}

// New creates and returns a new Bloom filter optimized for storing up to `n` items
// with a false positive probability of `p`.
//
// It returns an error if the probability is not in the range (0,1).
//
// Example:
//
//	bf, err := bitbloom.New(10000, 0.01)
//	if err != nil { log.Fatal(err) }
func New(n uint64, p float64) (*BloomFilter, error) {

	if p <= 0 || p >= 1 {
		return nil, fmt.Errorf("false positive rate must be 0 < p < 1")
	}

	m := OptimalM(n, p)
	k := OptimalK(m, n)
	return newBloomFilter(m, k), nil
}

// NewWithParams creates and returns a Bloom filter with explicit control over
// the size of the bit array (`m`) and number of hash functions (`k`).
//
// This should be used only if you need precise control over internals.
// For most users, the New() constructor is recommended.
func NewWithParams(m, k uint64) *BloomFilter {
	return newBloomFilter(m, k)
}

func newBloomFilter(m, k uint64) *BloomFilter {
	return &BloomFilter{
		bitset: bitset.New(m),
		hasher: hasher.New(),
		m:      m,
		k:      k,
	}
}

// Add inserts an item into the Bloom filter.
func (bf *BloomFilter) Add(item []byte) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	hashes := bf.hasher.Hashes(item, bf.k, bf.m)
	for _, h := range hashes {
		bf.bitset.Set(h)
	}

	bf.count++
}

// Test checks whether an item is possibly in the Bloom filter.
// Returns true if the item may be present (with false positives possible),
// or false if it is definitely not present.
func (bf *BloomFilter) Test(item []byte) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	hashes := bf.hasher.Hashes(item, bf.k, bf.m)
	for _, h := range hashes {
		if !bf.bitset.Get(h) {
			return false
		}
	}
	return true
}

// EstimatedFillRatio returns the theoretical fill ratio of the bit array
// based on the number of inserted elements and the number of hash functions.
func (bf *BloomFilter) EstimatedFillRatio() float64 {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	return 1 - math.Exp(-float64(bf.k*bf.count)/float64(bf.m))
}

// ActualFillRatio returns the real fill ratio (fraction of bits set)
// by counting the number of set bits in the bit array.
func (bf *BloomFilter) ActualFillRatio() float64 {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	setBits := bf.bitset.Count()
	return float64(setBits) / float64(bf.m)
}

// FalsePositiveRate estimates the current false positive rate
// based on the actual fill ratio and number of hash functions.
func (bf *BloomFilter) FalsePositiveRate() float64 {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	// (1 - e^(-k*n/m))^k ≈ (fillRatio)^k
	fillRatio := float64(bf.bitset.Count()) / float64(bf.m)
	return math.Pow(fillRatio, float64(bf.k))
}

// MemoryUsage returns the total memory used by the bit array in bytes.
func (bf *BloomFilter) MemoryUsage() int {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	words := (bf.m + 63) / 64
	return int(words) * 8
}

// MarshalBinary serializes the Bloom filter into a binary representation.
//
// The format of the serialized data is as follows (in little-endian order):
//
//	Offset  Size (bytes)  Description
//	------  ------------- ----------------------------------------------
//	0       8             m: total number of bits in the filter
//	8       8             k: number of hash functions used
//	16      8             count: number of items added
//	24      8 * w         bitset data (w = ceil(m / 64)) 64-bit words
//
// This binary encoding allows you to store or transmit the filter and
// restore it later using UnmarshalBinary. It is safe for cross-platform
// use as long as both sides use little-endian encoding.
//
// Example:
//
//	data, err := bf.MarshalBinary()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// save `data` to disk or send over network
//
// Returns a byte slice and any error encountered.
func (bf *BloomFilter) MarshalBinary() ([]byte, error) {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	words := (bf.m + 63) / 64
	buf := make([]byte, 24+words*8)

	binary.LittleEndian.PutUint64(buf[0:8], bf.m)
	binary.LittleEndian.PutUint64(buf[8:16], bf.k)
	binary.LittleEndian.PutUint64(buf[16:24], bf.count)

	bitsetData := bf.bitset.Data()
	for i, word := range bitsetData {
		offset := 24 + i*8
		binary.LittleEndian.PutUint64(buf[offset:offset+8], word)
	}

	return buf, nil
}

// UnmarshalBinary reconstructs a Bloom filter from its binary representation.
//
// The input byte slice must be in the format produced by MarshalBinary.
// It must contain at least 24 bytes of header followed by a valid bitset.
//
// Header format (little-endian):
//
//	Offset  Size (bytes)  Description
//	------  ------------- ----------------------------------------------
//	0       8             m: total number of bits in the filter
//	8       8             k: number of hash functions used
//	16      8             count: number of items added
//
// The remaining bytes must be the bitset data:
//
//	24      8 * w         bitset data (w = ceil(m / 64)) 64-bit words
//
// Validations performed:
//   - Ensures `m` and `k` are non-zero
//   - Ensures bitset data length matches expected word count
//   - Ensures bitset words are parsed correctly
//
// Example:
//
//	bf, err := bitbloom.UnmarshalBinary(data)
//	if err != nil {
//	    log.Fatal("Failed to deserialize filter:", err)
//	}
//
// Returns a new BloomFilter or an error if the data is invalid.
func UnmarshalBinary(data []byte) (*BloomFilter, error) {
	const headerSize = 24
	if len(data) < headerSize {
		return nil, fmt.Errorf("data too short for header")
	}

	m := binary.LittleEndian.Uint64(data[0:8])
	k := binary.LittleEndian.Uint64(data[8:16])
	count := binary.LittleEndian.Uint64(data[16:24])

	if m == 0 || k == 0 {
		return nil, fmt.Errorf("invalid parameters in serialized data")
	}

	bf := newBloomFilter(m, k)
	bf.count = count

	expectedWords := (m + 63) / 64
	actualWords := uint64(len(data[headerSize:])) / 8
	if actualWords != expectedWords {
		return nil, fmt.Errorf("bitset data length mismatch")
	}

	words := make([]uint64, expectedWords)
	for i := range words {
		words[i] = binary.LittleEndian.Uint64(data[headerSize+i*8:])
	}

	if err := bf.bitset.SetData(words); err != nil {
		return nil, fmt.Errorf("invalid bitset data: %w", err)
	}

	return bf, nil
}
