// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

func TestDigitMatch(t *testing.T) {
	a := assert.New(t)

	a.True(digitMatch("123"))
	a.True(digitMatch("0123"))
	a.False(digitMatch("0123x"))
	a.False(digitMatch("01.23"))
}

func TestWordMatch(t *testing.T) {
	a := assert.New(t)

	a.True(wordMatch("123"))
	a.True(wordMatch("a123"))
	a.True(wordMatch("Abc123"))
	a.True(wordMatch("Abc"))

	a.False(wordMatch("Ab c"))
	a.False(wordMatch("Ab_c"))
}
