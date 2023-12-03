//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package unit

import (
	"fmt"
	"math"
	"strconv"
)

const (
	Ki = float64(1024)
	Kb = float64(1000)
	Mi = Ki * 1024
	Mb = Kb * 1000
	Gi = Mi * 1024
	Gb = Mb * 1000
	Ti = Gi * 1024
	Tb = Gb * 1000
	Pi = Ti * 1024
	Pb = Tb * 1000
	Ei = Pi * 1024
	Eb = Pb * 1000
	Zi = Ei * 1024
	Zb = Eb * 1000
	Yi = Zi * 1024
	Yb = Zb * 1000
)

// BytesToSizeString takes a float64 number of bytes and returns a short size
// string, e.g. "102b", "5Ki", or "98Ti".
func BytesToSizeString(b float64) string {
	if math.Abs(b) < Ki {
		return fmt.Sprintf("%dB", int64(b))
	}
	b /= Ki
	for _, unit := range []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(b) < Ki {
			return fmt.Sprintf("%3.1f%s", b, unit)
		}
		b /= Ki
	}
	return fmt.Sprintf("%.1fYi", b)
}

// SizeStringToBytes returns the number of bytes given a size string such as
// "102", "3b", "5Ki", or "98Tb".
func SizeStringToBytes(s string) float64 {
	// if there's no suffix, just return the number...
	parsed, err := strconv.Atoi(s)
	if err == nil {
		return float64(parsed)
	}
	// find start of size suffix
	cur := 0
	for {
		c := s[cur]
		if c < '0' || c > '9' {
			break
		}
		cur++
	}
	baseInt, err := strconv.Atoi(s[0:cur])
	base := float64(baseInt)
	suffix := s[cur:len(s)]
	switch suffix {
	case "Ki":
		return base * Ki
	case "Kb":
		return base * Kb
	case "Mi":
		return base * Mi
	case "Mb":
		return base * Mb
	case "Gi":
		return base * Gi
	case "Gb":
		return base * Gb
	case "Ti":
		return base * Ti
	case "Tb":
		return base * Tb
	case "Pi":
		return base * Pi
	case "Pb":
		return base * Pb
	case "Ei":
		return base * Ei
	case "Eb":
		return base * Eb
	case "Zi":
		return base * Zi
	case "Zb":
		return base * Zb
	case "Yi":
		return base * Yi
	case "Yb":
		return base * Yb
	case "B":
		return base
	}
	return base
}
