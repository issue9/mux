// SPDX-License-Identifier: MIT

package syntax

type namedMatcher func(string) bool

var namedMatchers = map[string]namedMatcher{
	"any":   anyMatch,
	"digit": digitMatch,
	"word":  wordMatch,
}

func anyMatch(path string) bool { return true }

func digitMatch(path string) bool {
	for _, c := range path {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func wordMatch(path string) bool {
	for _, c := range path {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}
