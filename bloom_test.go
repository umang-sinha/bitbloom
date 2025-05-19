package bitbloom

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloomFilter_AddAndTest(t *testing.T) {
	bf, _ := New(1000, 0.01)

	item := []byte("golang")
	bf.Add(item)

	if !bf.Test(item) {
		t.Error("Expected item to be present in the filter after adding")
	}

	if bf.Test([]byte("python")) {
		t.Error("Unexpected item found in the filter")
	}
}

func TestBloomFilter_OptimalParams(t *testing.T) {
	m := OptimalM(1000, 0.01)
	k := OptimalK(m, 1000)

	if m == 0 || k == 0 {
		t.Error("Expected non-zero m and k for optimal parameters")
	}
}

func TestBloomFilter_NewWithParams(t *testing.T) {
	bf := NewWithParams(1024, 3)

	if bf == nil {
		t.Fatal("Expected non-nil BloomFilter instance")
	}
}

func TestBloomFilter_EstimatedFillRatio(t *testing.T) {
	bf := NewWithParams(1000, 3)

	for i := 0; i < 100; i++ {
		bf.Add([]byte{byte(i)})
	}

	ratio := bf.EstimatedFillRatio()
	if ratio <= 0 || ratio >= 1 {
		t.Error("Expected EstimatedFillRatio to be between 0 and 1")
	}
}

func TestBloomFilter_ActualFillRatio(t *testing.T) {
	bf := NewWithParams(1000, 3)

	bf.Add([]byte("foo"))
	ratio := bf.ActualFillRatio()

	if ratio <= 0 {
		t.Error("Expected some bits to be set after adding an item")
	}
}

func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	bf := NewWithParams(10000, 4)

	for i := 0; i < 1000; i++ {
		bf.Add([]byte{byte(i)})
	}

	rate := bf.FalsePositiveRate()
	if rate <= 0 || rate >= 1 {
		t.Error("Expected false positive rate to be between 0 and 1")
	}
}

func TestBloomFilter_MemoryUsage(t *testing.T) {
	bf := NewWithParams(1000, 3)
	mem := bf.MemoryUsage()

	if mem <= 0 {
		t.Error("Expected positive memory usage")
	}
}

func TestBloomFilter_MarshalUnmarshal(t *testing.T) {
	bf := NewWithParams(1000, 3)
	bf.Add([]byte("foo"))
	bf.Add([]byte("bar"))

	data, err := bf.MarshalBinary()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	newBf, err := UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !newBf.Test([]byte("foo")) || !newBf.Test([]byte("bar")) {
		t.Error("Unmarshalled Bloom filter should contain the original items")
	}
}

func TestBloomFilter_UnmarshalInvalidData(t *testing.T) {
	_, err := UnmarshalBinary([]byte("short"))
	if err == nil {
		t.Error("Expected error when unmarshalling short data")
	}

	// Invalid m/k
	buf := make([]byte, 24)
	data := append(buf, make([]byte, 8)...)
	_, err = UnmarshalBinary(data)
	if err == nil {
		t.Error("Expected error for zero m and k")
	}
}

func TestBloomFilter_TestBeforeAdd(t *testing.T) {
	bf, _ := New(100, 0.01)
	item := []byte("ghost")
	if bf.Test(item) {
		t.Errorf("Expected item to be absent before adding")
	}
}

func TestBloomFilter_RepeatedAdd(t *testing.T) {
	bf, _ := New(100, 0.01)
	item := []byte("repeat")
	for i := 0; i < 1000; i++ {
		bf.Add(item)
	}
	if !bf.Test(item) {
		t.Errorf("Item should still be found after repeated additions")
	}
}

func TestBloomFilter_MarshalUnmarshal_Empty(t *testing.T) {
	bf, _ := New(100, 0.01)
	data, err := bf.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	newBF, err := UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}
	if newBF.count != 0 {
		t.Errorf("Expected count 0 after unmarshalling empty filter, got %d", newBF.count)
	}
}

func TestBloomFilter_UnmarshalBinary_InvalidHeader(t *testing.T) {
	_, err := UnmarshalBinary([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("Expected error for invalid header size")
	}
}

func TestBloomFilter_UnmarshalBinary_CorruptBitset(t *testing.T) {
	bf, _ := New(100, 0.01)
	data, _ := bf.MarshalBinary()
	data = data[:len(data)-8] // truncate to simulate corruption

	_, err := UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Expected error for corrupted bitset")
	}
}

func TestBloomFilter_ConcurrentAdd(t *testing.T) {
	bf, _ := New(10000, 0.01)
	var wg sync.WaitGroup
	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item := []byte(fmt.Sprintf("item-%d", i))
			bf.Add(item)
		}(i)
	}
	wg.Wait()

	count := 0
	for i := 0; i < n; i++ {
		item := []byte(fmt.Sprintf("item-%d", i))
		if bf.Test(item) {
			count++
		}
	}

	if count < n {
		t.Errorf("Expected at least %d positives, got %d", n, count)
	}
}

func TestBloomFilter_ConcurrentTest(t *testing.T) {
	bf, _ := New(10000, 0.01)
	for i := 0; i < 1000; i++ {
		bf.Add([]byte(fmt.Sprintf("item-%d", i)))
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				bf.Test([]byte(fmt.Sprintf("item-%d", j)))
			}
		}()
	}
	wg.Wait()
}

func TestBloomFilter_ConcurrentAddAndTest(t *testing.T) {
	bf, _ := New(10000, 0.01)
	var wg sync.WaitGroup
	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			bf.Add([]byte(fmt.Sprintf("item-%d", i)))
		}(i)

		go func(i int) {
			defer wg.Done()
			bf.Test([]byte(fmt.Sprintf("item-%d", i)))
		}(i)
	}
	wg.Wait()
}

func TestBloomFilter_ConcurrentFillRatios(t *testing.T) {
	bf, _ := New(10000, 0.01)
	for i := 0; i < 1000; i++ {
		bf.Add([]byte(fmt.Sprintf("item-%d", i)))
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bf.EstimatedFillRatio()
			_ = bf.ActualFillRatio()
			_ = bf.FalsePositiveRate()
		}()
	}
	wg.Wait()
}

func TestBloomFilter_ConcurrentMarshal(t *testing.T) {
	bf, _ := New(10000, 0.01)
	for i := 0; i < 1000; i++ {
		bf.Add([]byte(fmt.Sprintf("item-%d", i)))
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := bf.MarshalBinary()
			if err != nil || len(data) == 0 {
				t.Errorf("Failed to marshal bloom filter: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestBloomFilter_AddDuringMarshal(t *testing.T) {
	bf, _ := New(10000, 0.01)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			bf.Add([]byte(fmt.Sprintf("item-%d", i)))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_, err := bf.MarshalBinary()
			if err != nil {
				t.Errorf("Marshal failed during concurrent writes: %v", err)
			}
		}
	}()

	wg.Wait()
}

func TestBloomFilter_AllOpsConcurrent(t *testing.T) {
	bf, _ := New(10000, 0.01)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(4)

		go func(i int) {
			defer wg.Done()
			bf.Add([]byte(fmt.Sprintf("item-%d", i)))
		}(i)

		go func(i int) {
			defer wg.Done()
			bf.Test([]byte(fmt.Sprintf("item-%d", i)))
		}(i)

		go func() {
			defer wg.Done()
			_ = bf.EstimatedFillRatio()
		}()

		go func() {
			defer wg.Done()
			_, _ = bf.MarshalBinary()
		}()
	}

	wg.Wait()
}

func TestBloomFilter_ConcurrentReadOnly(t *testing.T) {
	bf, _ := New(5000, 0.01)

	for i := 0; i < 500; i++ {
		bf.Add([]byte(fmt.Sprintf("word-%d", i)))
	}

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bf.EstimatedFillRatio()
			_ = bf.ActualFillRatio()
			_ = bf.FalsePositiveRate()
			_ = bf.MemoryUsage()
		}()
	}

	wg.Wait()
}

func TestBloomFilter_StressConcurrentAccess(t *testing.T) {
	bf, _ := New(100000, 0.01)
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item := []byte(fmt.Sprintf("data-%d", i))
			bf.Add(item)
			assert.True(t, bf.Test(item))
			_ = bf.EstimatedFillRatio()
			_ = bf.FalsePositiveRate()
		}(i)
	}

	wg.Wait()
}

func TestBloomFilter_ConcurrentUnmarshalAndTest(t *testing.T) {
	original, _ := New(1000, 0.01)
	for i := 0; i < 100; i++ {
		original.Add([]byte(fmt.Sprintf("word-%d", i)))
	}
	data, _ := original.MarshalBinary()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = UnmarshalBinary(data)
		}()
		go func() {
			defer wg.Done()
			original.Test([]byte("word-1"))
		}()
	}

	wg.Wait()
}
