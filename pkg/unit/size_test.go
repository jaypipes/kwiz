// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package unit_test

import (
	"testing"

	"github.com/jaypipes/kwiz/pkg/unit"
)

func TestBytesToSizeString(t *testing.T) {
	tcs := []struct {
		val float64
		exp string
	}{
		{float64(1021), "1021B"},
		{float64(1024), "1.0Ki"},
		{float64(1024 * 1024), "1.0Mi"},
		{float64(64 * 1024 * 1024), "64.0Mi"},
		{float64(6.7108864e+07), "64.0Mi"},
	}

	for _, tc := range tcs {
		got := unit.BytesToSizeString(tc.val)
		if got != tc.exp {
			t.Fatalf("expected %s but got %s", tc.exp, got)
		}
	}
}

func TestSizeStringToBytes(t *testing.T) {
	tcs := []struct {
		val string
		exp float64
	}{
		{"1021", float64(1021)},
		{"12B", float64(12)},
		{"1Ki", float64(1024)},
		{"1Mi", float64(1024 * 1024)},
		{"64Mi", float64(64 * 1024 * 1024)},
		{"64Mb", float64(64 * 1000 * 1000)},
		{"128Gi", float64(128 * 1024 * 1024 * 1024)},
	}

	for _, tc := range tcs {
		got := unit.SizeStringToBytes(tc.val)
		if got != tc.exp {
			t.Fatalf("expected %3.1f but got %3.1f", tc.exp, got)
		}
	}
}
