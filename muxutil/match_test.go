// SPDX-License-Identifier: MIT

package muxutil

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7/types"
)

func TestAndMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := AndMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps := types.NewContext("")
	ok := m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/path")

	m = AndMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps = types.NewContext("")
	ok = m.Match(r, ps)
	a.False(ok)
}

func TestOrMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := OrMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps := types.NewContext("")
	ok := m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/v2/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps = types.NewContext("")
	ok = m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/v1/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v111/v2/v1/path").Request()
	ps = types.NewContext("")
	ok = m.Match(r, ps)
	a.False(ok)
}
