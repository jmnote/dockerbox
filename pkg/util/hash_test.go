package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	testcases := []struct {
		input    string
		length   int
		wantHash string
	}{
		{"", 5, "bwgu8"},
		{"", 10, "bwgu80skaz"},
		{"", 100, "bwgu80skazsk93503utcsb34k59rca4dbwgu80skazsk93503utcsb34k59rca4dbwgu80skazsk93503utcsb34k59rca4dbwgu"},
		{"hello", 5, "8q56n"},
		{"hello", 10, "8q56nwje2g"},
		{"hello", 100, "8q56nwje2gn6h5aermukvnum74fq3v808q56nwje2gn6h5aermukvnum74fq3v808q56nwje2gn6h5aermukvnum74fq3v808q56"},
		{"world", 5, "02kq0"},
		{"world", 10, "02kq0t772k"},
		{"world", 100, "02kq0t772kr7ga1yz0ksa7iyi0t2ec4n02kq0t772kr7ga1yz0ksa7iyi0t2ec4n02kq0t772kr7ga1yz0ksa7iyi0t2ec4n02kq"},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			got := Hash(tc.input, tc.length)
			require.Equal(t, tc.wantHash, got)
		})
	}
}

func TestNewHash(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			h := NewHash(i)
			require.Equal(t, i, len(h))
		})
	}
}
