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
	var out [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		if i >= 3 {
			break
		}
		out[i], _ = strconv.Atoi(s)
	}
	return out
}
