# bitbloom

[![Go Reference](https://pkg.go.dev/badge/github.com/umang-sinha/bitbloom.svg)](https://pkg.go.dev/github.com/umang-sinha/bitbloom)

**bitbloom** is a high-performance, thread-safe Bloom filter implementation in Go.

A Bloom filter is a probabilistic data structure used to test whether an element is a member of a set.
False positives are possible, but false negatives are not. This means that it can tell you with certainty
that an element is not in the set, but it may incorrectly report that an element is present.

## Features

- **Optimal Parameter Calculation:** Automatically calculates the optimal bit array size and number of hash functions based on the expected number of items and desired false positive probability.
- **Thread-Safe:** Designed for concurrent use, ensuring safe and efficient `Add` and `Test` operations from multiple goroutines.
- **Fill Ratio Estimation:** Provides methods to estimate and track the fill ratio of the Bloom filter, helping in performance analysis and tuning.
- **Memory Efficient:** Uses a compact bitset for efficient storage.
- **Serialization:** Supports binary serialization for persistence or data transfer, allowing you to save and restore the filter state.

## Installation

To install bitbloom, use `go get`:

```bash
go get github.com/umang-sinha/bitbloom
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/umang-sinha/bitbloom"
)

func main() {
	// Create a new Bloom filter that can store 1 million items with a 1% false positive rate
	bf, err := bitbloom.New(1000000, 0.01)
	if err != nil {
		log.Fatal(err)
	}

	// Add an item to the filter
	bf.Add([]byte("golang"))

	// Test if the item is possibly present in the filter
	if bf.Test([]byte("golang")) {
		fmt.Println("Item possibly present")
	} else {
		fmt.Println("Item is definitely not present")
	}

	// Example of a false positive (rare, but possible)
	if bf.Test([]byte("python")) {
		fmt.Println("Item possibly present (false positive)")
	} else {
		fmt.Println("Item is definitely not present")
	}
}
```

## API Reference

- ```New(n uint64, p float64) (*BloomFilter, error)```

Creates a new Bloom filter.

    n: Expected number of items to be stored.
    p: Desired false positive probability (between 0 and 1).

Returns a new BloomFilter or an error if the probability is invalid.

- ```NewWithParams(m uint64, k uint64) *BloomFilter```

Creates a new Bloom filter with explicit parameters.

    m: Size of the bit array.
    k: Number of hash functions.

Use this for fine-grained control; otherwise, ```New()``` is recommended.

- ```(*BloomFilter) Add(item []byte)```

Adds an item to the Bloom filter.

- ```(*BloomFilter) Test(item []byte) bool```

Checks if an item is possibly present in the Bloom filter.  Returns true if the item might be present (false positive possible), and false if it is definitely not present.

- ```(*BloomFilter) EstimatedFillRatio() float64```

Returns the theoretical fill ratio of the Bloom filter.

- ```(*BloomFilter) ActualFillRatio() float64```

Returns the actual fill ratio (fraction of bits set) of the Bloom filter.

- ```(*BloomFilter) FalsePositiveRate() float64```

Estimates the current false positive rate.

- ```(*BloomFilter) MemoryUsage() int```

Returns the memory usage of the Bloom filter in bytes.

- ```(*BloomFilter) MarshalBinary() ([]byte, error)```

Serializes the Bloom filter to a binary format.  This is useful for saving the filter or sending it over a network.

- ```UnmarshalBinary(data []byte) (*BloomFilter, error)```

Deserializes a Bloom filter from its binary representation.

- ```OptimalM(n uint64, p float64) uint64```

Calculates the optimal size of the bit array (m).

- ```OptimalK(m uint64, n uint64) uint64```

Calculates the optimal number of hash functions (k).

## Thread Safety

**bitbloom** is thread-safe.  Multiple goroutines can safely call ```Add``` and ```Test``` concurrently.  Internal locking mechanisms ensure data consistency.

## Performance

bitbloom is designed for high performance.  It uses an efficient bitset implementation and the fast MurmurHash3 hashing algorithm.  The optimal parameter calculation helps to minimize memory usage and false positive rates.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the [MIT licence](https://github.com/umang-sinha/bitbloom/blob/main/LICENSE).