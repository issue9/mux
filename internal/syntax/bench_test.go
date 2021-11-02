// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/params"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	p := &MatchParam{
		Path: "100000/author",
	}
	for i := 0; i < b.N; i++ {
		index := seg.Match(p)
		a.Equal(index, len(p.Path)).
			Equal(p.Params, map[string]string{"id": "100000"})
	}
}

func BenchmarkSegment_Match_Named_withMatcher(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	p := &MatchParam{
		Path: "10000/author",
	}
	for i := 0; i < b.N; i++ {
		index := seg.Match(p)
		a.Equal(index, len(p.Path)).
			Equal(p.Params, params.Params{"id": "10000"})
	}
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	p := &MatchParam{
		Path: "/posts/author",
	}
	for i := 0; i < b.N; i++ {
		index := seg.Match(p)
		a.Equal(index, len(p.Path)).Nil(p.Params)
	}
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	p := &MatchParam{
		Path: "1/author",
	}
	for i := 0; i < b.N; i++ {
		index := seg.Match(p)
		a.Equal(index, len(p.Path)).
			Equal(p.Params, params.Params{"id": "1"})
	}
}
