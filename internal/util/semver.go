package util

import (
	"strconv"
	"strings"
)

// SemverGreater returns true if version a is greater than b using major.minor.patch comparison.
func SemverGreater(a, b string) bool {
	ap := parseSemver(a)
	bp := parseSemver(b)
	for i := range ap {
		if ap[i] > bp[i] {
			return true
		}
		if ap[i] < bp[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	var out [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		if i >= 3 {
			break
		}
		// strip pre-release/build metadata (e.g. "35-3-g70ed907-dirty" → "35")
		if idx := strings.IndexByte(s, '-'); idx >= 0 {
			s = s[:idx]
		}
		out[i], _ = strconv.Atoi(s)
	}
	return out
}
