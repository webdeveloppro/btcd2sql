package db2sql

import (
	"testing"
)

// TestGetInputAddress will test if we have a right hex to bitcoin address translator
func TestGetInputAddress(t *testing.T) {
	tables := []struct {
		hex string
		out string
		e   string
	}{
		{
			"044f64882e2ac124513d9934b9dd5e62ec49c8e838374e401fda7225a05c8944405877231c4b68dadd5177c98b94bf09ed238591cdc44415a3a4e09ddca45d6d13",
			"1Gxg4bXL2qGmCeeGFboKGma86bTj5U1XxA",
			"",
		},
		{
			"123",
			"",
			"encoding/hex: odd length hex string",
		},
	}

	for _, test := range tables {
		address, err := GetInputAddress(test.hex)
		if err != nil {
			if err.Error() != test.e {
				t.Errorf("Expected error value do not match, got: %v, expected: %v", err, test.e)
			}
		}
		if address != test.out {
			t.Errorf("GetInputAddress output value do not match: %s, want: %s.", address, test.out)
		}
	}
}
