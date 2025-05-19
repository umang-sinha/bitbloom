package bitset

import (
	"testing"
)

func TestBitSet_SetAndGet(t *testing.T) {
	bs := New(128)

	bs.Set(5)
	if !bs.Get(5) {
		t.Errorf("Bit at position 5 should be set")
	}

	if bs.Get(6) {
		t.Errorf("Bit at position 6 should not be set")
	}
}

func TestBitSet_OutOfBounds(t *testing.T) {
	bs := New(64)

	bs.Set(1000) // should not panic
	if bs.Get(1000) {
		t.Errorf("Bit out of bounds should return false")
	}
}

func TestBitSet_MultipleSets(t *testing.T) {
	bs := New(128)

	positions := []uint64{0, 15, 31, 63, 64, 65, 127}
	for _, pos := range positions {
		bs.Set(pos)
	}

	for _, pos := range positions {
		if !bs.Get(pos) {
			t.Errorf("Bit at position %d should be set", pos)
		}
	}
}

func TestBitSet_RepeatSet(t *testing.T) {
	bs := New(64)
	bs.Set(10)
	bs.Set(10)
	bs.Set(10)

	if !bs.Get(10) {
		t.Errorf("Bit at 10 should remain set after multiple sets")
	}

	if bs.Count() != 1 {
		t.Errorf("Count should be 1, got %d", bs.Count())
	}
}

func TestBitSet_AllUnsetInitially(t *testing.T) {
	bs := New(100)
	for i := uint64(0); i < 100; i++ {
		if bs.Get(i) {
			t.Errorf("Bit at position %d should be unset initially", i)
		}
	}
}

func TestBitSet_AllSet(t *testing.T) {
	bs := New(100)
	for i := uint64(0); i < 100; i++ {
		bs.Set(i)
	}

	for i := uint64(0); i < 100; i++ {
		if !bs.Get(i) {
			t.Errorf("Bit at position %d should be set", i)
		}
	}

	if bs.Count() != 100 {
		t.Errorf("Expected count to be 100, got %d", bs.Count())
	}
}

func TestBitSet_Size(t *testing.T) {
	size := uint64(73)
	bs := New(size)
	if bs.Size() != size {
		t.Errorf("Expected size to be %d, got %d", size, bs.Size())
	}
}

func TestBitSet_SetEdgeBits(t *testing.T) {
	bs := New(128)

	edgePositions := []uint64{0, 63, 64, 127}
	for _, pos := range edgePositions {
		bs.Set(pos)
	}

	for _, pos := range edgePositions {
		if !bs.Get(pos) {
			t.Errorf("Bit at edge position %d should be set", pos)
		}
	}
}

func TestBitSet_CountAfterRandomSets(t *testing.T) {
	bs := New(1000)

	positions := []uint64{0, 10, 20, 30, 40, 500, 999}
	for _, pos := range positions {
		bs.Set(pos)
	}

	if bs.Count() != uint(len(positions)) {
		t.Errorf("Expected count %d, got %d", len(positions), bs.Count())
	}
}

func TestBitSet_SparseAndDenseSets(t *testing.T) {
	bs := New(1024)

	for i := uint64(0); i < 1024; i += 100 {
		bs.Set(i)
	}

	expected := uint(11)
	if bs.Count() != expected {
		t.Errorf("Expected sparse count %d, got %d", expected, bs.Count())
	}

	bs2 := New(256)
	for i := uint64(0); i < 256; i++ {
		bs2.Set(i)
	}

	if bs2.Count() != 256 {
		t.Errorf("Expected dense count 256, got %d", bs2.Count())
	}
}

func TestBitSet_InterleavedSets(t *testing.T) {
	bs := New(128)

	for i := uint64(0); i < 128; i += 2 {
		bs.Set(i)
	}

	for i := uint64(0); i < 128; i++ {
		expected := i%2 == 0
		if bs.Get(i) != expected {
			t.Errorf("Expected Get(%d) to be %v", i, expected)
		}
	}
}

func TestBitSet_LargeIndexHandling(t *testing.T) {
	bs := New(1 << 20) // 1M bits
	bs.Set(1 << 19)    // Set halfway
	if !bs.Get(1 << 19) {
		t.Errorf("Expected bit at 2^19 to be set")
	}
	if bs.Get(1 << 20) {
		t.Errorf("Expected bit at 2^20 to be out of bounds and false")
	}
}

func TestBitSet_SetResetGet(t *testing.T) {
	bs := New(64)
	bs.Set(32)
	if !bs.Get(32) {
		t.Errorf("Bit 32 should be set")
	}
	bs.data[32/64] &^= (1 << (32 % 64)) // Manually unset bit for testing
	if bs.Get(32) {
		t.Errorf("Bit 32 should be unset after manual clear")
	}
}

func TestBitSet_OverflowSafe(t *testing.T) {
	bs := New(5)
	bs.Set(1000)
	bs.Set(99999)
	bs.Set(^uint64(0)) // max uint64
	if bs.Get(1000) || bs.Get(99999) || bs.Get(^uint64(0)) {
		t.Errorf("Out-of-bound set/get should be ignored and return false")
	}
}
