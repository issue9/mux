// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ Matcher = MatcherFunc(Any)

func TestAny(t *testing.T) {
	a := assert.New(t)

	r, ok := Any(nil)
	a.True(ok).Nil(r)

	r, ok = MatcherFunc(Any).Match(nil)
	a.True(ok).Nil(r)
}

func TestAndFunc(t *testing.T) {
	a := assert.New(t)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := AndFunc(p1.Match, p2.Match)
	r := httptest.NewRequest(http.MethodGet, "/v1/v2/path", nil)
	rr, ok := m.Match(r)
	a.True(ok).NotNil(rr).
		Equal(rr.URL.Path, "/path")

	m = AndFunc(p1.Match, p2.Match)
	r = httptest.NewRequest(http.MethodGet, "/v2/v1/path", nil)
	rr, ok = m.Match(r)
	a.False(ok).Nil(rr)
}

func TestOrFunc(t *testing.T) {
	a := assert.New(t)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := OrFunc(p1.Match, p2.Match)
	r := httptest.NewRequest(http.MethodGet, "/v1/v2/path", nil)
	rr, ok := m.Match(r)
	a.True(ok).NotNil(rr).
		Equal(rr.URL.Path, "/v2/path")

	m = OrFunc(p1.Match, p2.Match)
	r = httptest.NewRequest(http.MethodGet, "/v2/v1/path", nil)
	rr, ok = m.Match(r)
	a.True(ok).NotNil(rr).
		Equal(rr.URL.Path, "/v1/path")

	m = OrFunc(p1.Match, p2.Match)
	r = httptest.NewRequest(http.MethodGet, "/v111/v2/v1/path", nil)
	rr, ok = m.Match(r)
	a.False(ok).Nil(rr)
}
