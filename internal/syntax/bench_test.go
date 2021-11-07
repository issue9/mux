// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/params"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := &MatchParam{Path: "100000/author"}
		a.True(seg.Match(p)).Equal(p.Params, map[string]string{"id": "100000"})
	}
}

func BenchmarkSegment_Match_Named_withMatcher(b *testing.B) {
	a := assert.New(b)
	i := NewInterceptors()
	i.Add(MatchDigit, "digit")
	a.NotNil(i)

	seg, err := i.NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := &MatchParam{Path: "10000/author"}
		a.True(seg.Match(p)).Equal(p.Params, params.Params{"id": "10000"})
	}
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := &MatchParam{Path: "/posts/author"}
		a.True(seg.Match(p)).Nil(p.Params)
	}
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	for i := 0; i < b.N; i++ {
		p := &MatchParam{Path: "1/author"}
		a.True(seg.Match(p)).Equal(p.Params, params.Params{"id": "1"})
	}
}
