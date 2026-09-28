// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v5"

	"github.com/issue9/mux/v10/types"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	route := types.NewRoute()
	var ok bool
	for b.Loop() {
		route.Reset()
		route.Path = "100000/author"
		ok = seg.Match(route)
	}
	a.True(ok).Equal(route.MustString("id", "not-exits"), "100000") // 相同的数据，仅判断最后一次数据
}

func BenchmarkSegment_Match_Named_withInterceptors(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	i.Add(MatchDigit, "digit")
	a.NotNil(i)

	seg, err := i.NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	route := types.NewRoute()
	var ok bool
	for b.Loop() {
		route.Reset()
		route.Path = "100000/author"
		ok = seg.Match(route)
	}
	a.True(ok).Equal(route.MustString("id", "not-exits"), "100000")
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	route := types.NewRoute()
	var ok bool
	for b.Loop() {
		route.Reset()
		route.Path = "/posts/author"
		ok = seg.Match(route)
	}
	a.True(ok).Zero(route.Count())
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b, false)
	i := NewInterceptors()
	a.NotNil(i)

	seg, err := i.NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	route := types.NewRoute()
	var ok bool
	for b.Loop() {
		route.Reset()
		route.Path = "1/author"
		ok = seg.Match(route)
	}
	a.True(ok).Equal(route.MustString("id", "not-exists"), "1")
}
