// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func newInterceptors(a *assert.Assertion) *Interceptors {
	i := NewInterceptors()
	a.NotNil(i)

	i.Add(MatchDigit, "digit")
	i.Add(MatchWord, "word")
	i.Add(MatchAny, "any")

	return i
}

func TestInterceptors_Add(t *testing.T) {
	a := assert.New(t, false)
	i := NewInterceptors()

	a.Panic(func() {
		i.Add(MatchAny)
	})

	i.Add(MatchWord, "[a-zA-Z0-9]+", "word1", "word2")
	a.PanicString(func() {
		i.Add(MatchWord, "[a-zA-Z0-9]+")
	}, "已经存在")
	_, found := i.funcs["word1"]
	a.True(found)
	_, found = i.funcs["[a-zA-Z0-9]+"]
	a.True(found)
}

func TestMatchAny(t *testing.T) {
	a := assert.New(t, false)

	a.True(MatchAny("1"))
	a.True(MatchAny("_"))
	a.True(MatchAny("."))
	a.True(MatchAny(" "))
	a.True(MatchAny("\t"))

	a.False(MatchAny(""))
}

func TestMatchDigit(t *testing.T) {
	a := assert.New(t, false)

	a.True(MatchDigit("123"))
	a.True(MatchDigit("0123"))

	a.False(MatchDigit("0123x"))
	a.False(MatchDigit("01.23"))
	a.False(MatchDigit(""))
}

func TestMatchWord(t *testing.T) {
	a := assert.New(t, false)

	a.True(MatchWord("123"))
	a.True(MatchWord("a123"))
	a.True(MatchWord("Abc123"))
	a.True(MatchWord("Abc"))

	a.False(MatchWord("Ab c"))
	a.False(MatchWord("Ab_c"))
	a.False(MatchWord(""))
}
