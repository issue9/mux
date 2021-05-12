// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/params"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	ps := params.Params{}
	path := "100000/author"
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, ps)
		a.Equal(index, len(path))
	}
}

func BenchmarkSegment_Match_Named_withMatcher(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	ps := params.Params{}
	path := "10000/author"
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, ps)
		a.Equal(index, len(path))
	}
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	path := "/posts/author"
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, nil)
		a.Equal(index, len(path))
	}
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b)

	seg, err := NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	path := "1/author"
	ps := params.Params{}
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, ps)
		a.Equal(index, len(path))
	}
}
