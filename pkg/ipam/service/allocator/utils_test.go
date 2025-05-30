// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium
// Copyright The Kubernetes Authors.

package allocator

import (
	"math/big"
	"testing"
)

func TestCountBits(t *testing.T) {
	// bigN is an integer that occupies more than one big.Word.
	bigN, ok := big.NewInt(0).SetString("10000000000000000000000000000000000000000000000000000000000000000", 16)
	if !ok {
		t.Fatal("Failed to set bigN")
	}
	tests := []struct {
		n        *big.Int
		expected int
	}{
		{n: big.NewInt(int64(0)), expected: 0},
		{n: big.NewInt(int64(0xffffffffff)), expected: 40},
		{n: bigN, expected: 1},
	}
	for _, test := range tests {
		actual := countBits(test.n)
		if test.expected != actual {
			t.Errorf("%s should have %d bits but recorded as %d", test.n, test.expected, actual)
		}
	}
}

func BenchmarkCountBits(b *testing.B) {
	bigN, ok := big.NewInt(0).SetString("10000000000000000000000000000000000000000000000000000000000000000", 16)
	if !ok {
		b.Fatal("Failed to set bigN")
	}
	for b.Loop() {
		countBits(bigN)
	}
}
