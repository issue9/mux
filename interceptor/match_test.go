// SPDX-License-Identifier: MIT

package interceptor

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMatchDigit(t *testing.T) {
	a := assert.New(t)

	a.True(MatchDigit("123"))
	a.True(MatchDigit("0123"))

	a.False(MatchDigit("0123x"))
	a.False(MatchDigit("01.23"))
	a.False(MatchDigit(""))
}

func TestMatchWord(t *testing.T) {
	a := assert.New(t)

	a.True(MatchWord("123"))
	a.True(MatchWord("a123"))
	a.True(MatchWord("Abc123"))
	a.True(MatchWord("Abc"))

	a.False(MatchWord("Ab c"))
	a.False(MatchWord("Ab_c"))
	a.False(MatchWord(""))
}
