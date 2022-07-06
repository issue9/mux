// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v7/types"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := types.NewContext("100000/author")
		a.True(seg.Match(p)).Equal(p.MustString("id", "not-exits"), "100000")
	}
}

func BenchmarkSegment_Match_Named_withMatcher(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	i.Add(MatchDigit, "digit")
	a.NotNil(i)

	seg, err := i.NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := types.NewContext("100000/author")
		a.True(seg.Match(p)).Equal(p.MustString("id", "not-exits"), "100000")
	}
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := types.NewContext("/posts/author")
		a.True(seg.Match(p)).Zero(p.Count())
	}
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := types.NewContext("1/author")
		a.True(seg.Match(p)).Equal(p.MustString("id", "not-exists"), "1")
	}
}
