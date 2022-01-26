// SPDX-License-Identifier: MIT

package muxutil

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func TestAndMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := AndMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps, ok := m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/path")

	m = AndMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.False(ok).Nil(ps)
}

func TestOrMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := OrMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps, ok := m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/v2/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/v1/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v111/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.False(ok).Nil(ps)
}
