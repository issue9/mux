// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

var _ Matcher = MatcherFunc(Any)

func TestAny(t *testing.T) {
	a := assert.New(t, false)

	r, ok := Any(nil)
	a.True(ok).Nil(r)

	r, ok = MatcherFunc(Any).Match(nil)
	a.True(ok).Nil(r)
}

func TestAndFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := AndFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps, ok := m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/path")

	m = AndFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.False(ok).Nil(ps)
}

func TestOrFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := OrFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps, ok := m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/v2/path")

	m = OrFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.True(ok).NotNil(ps).
		Equal(r.URL.Path, "/v1/path")

	m = OrFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v111/v2/v1/path").Request()
	ps, ok = m.Match(r)
	a.False(ok).Nil(ps)
}
