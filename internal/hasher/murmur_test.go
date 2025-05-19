package hasher

import (
	"encoding/hex"
	"testing"
)

func TestMurmurHasher_ConsistentHashing(t *testing.T) {
	mh := New()

	data := []byte("test input")
	k, m := uint64(5), uint64(100)
	hashes1 := mh.Hashes(data, k, m)
	hashes2 := mh.Hashes(data, k, m)

	if len(hashes1) != int(k) || len(hashes2) != int(k) {
		t.Fatalf("Expected %d hashes, got %d and %d", k, len(hashes1), len(hashes2))
	}

	for i := range hashes1 {
		if hashes1[i] != hashes2[i] {
			t.Errorf("Hash mismatch at index %d: %d != %d", i, hashes1[i], hashes2[i])
		}
	}
}

func TestMurmurHasher_DifferentInputsProduceDifferentHashes(t *testing.T) {
	mh := New()

	data1 := []byte("hello")
	data2 := []byte("world")

	hashes1 := mh.Hashes(data1, 4, 100)
	hashes2 := mh.Hashes(data2, 4, 100)

	// High chance at least one differs
	sameCount := 0
	for i := 0; i < 4; i++ {
		if hashes1[i] == hashes2[i] {
			sameCount++
		}
	}

	if sameCount == 4 {
		t.Errorf("Expected different hashes, but got same hashes for all k values")
	}
}

func TestMurmurHasher_KZero(t *testing.T) {
	mh := New()

	hashes := mh.Hashes([]byte("data"), 0, 100)
	if len(hashes) != 0 {
		t.Errorf("Expected 0 hashes for k=0, got %d", len(hashes))
	}
}

func TestMurmurHasher_MZero(t *testing.T) {
	mh := New()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic or error when m = 0")
		}
	}()

	_ = mh.Hashes([]byte("data"), 3, 0) // modulo by zero
}

func TestMurmurHasher_RepeatedCallsWithSameInput(t *testing.T) {
	mh := New()
	data := []byte("repeat-test")
	k, m := uint64(10), uint64(1024)

	expected := mh.Hashes(data, k, m)
	for i := 0; i < 10; i++ {
		actual := mh.Hashes(data, k, m)
		for j := range expected {
			if expected[j] != actual[j] {
				t.Errorf("Mismatch on iteration %d, index %d: expected %d, got %d", i, j, expected[j], actual[j])
			}
		}
	}
}

func TestMurmurHasher_LargeK(t *testing.T) {
	mh := New()
	k, m := uint64(10000), uint64(1_000_000)
	hashes := mh.Hashes([]byte("large-k-test"), k, m)

	if len(hashes) != int(k) {
		t.Errorf("Expected %d hashes, got %d", k, len(hashes))
	}
	for _, h := range hashes {
		if h >= m {
			t.Errorf("Hash value %d out of range (>= %d)", h, m)
		}
	}
}

func TestMurmurHasher_EmptyData(t *testing.T) {
	mh := New()

	hashes := mh.Hashes([]byte{}, 5, 128)
	if len(hashes) != 5 {
		t.Errorf("Expected 5 hashes for empty input, got %d", len(hashes))
	}
}

func TestMurmurHasher_HexEncodedInput(t *testing.T) {
	mh := New()

	rawData, _ := hex.DecodeString("deadbeef")
	hashes := mh.Hashes(rawData, 3, 256)

	if len(hashes) != 3 {
		t.Errorf("Expected 3 hashes, got %d", len(hashes))
	}
}

func TestMurmurHasher_HashDiversity(t *testing.T) {
	mh := New()

	data1 := []byte("collision")
	data2 := []byte("collision!")

	k := uint64(5)
	m := uint64(64)
	hashes1 := mh.Hashes(data1, k, m)
	hashes2 := mh.Hashes(data2, k, m)

	collisions := 0
	for i := 0; i < int(k); i++ {
		if hashes1[i] == hashes2[i] {
			collisions++
		}
	}

	// Allowing at most 2 collisions between slightly different inputs
	if collisions > 2 {
		t.Errorf("Too many hash collisions: got %d collisions out of %d", collisions, k)
	}
}
