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

	a.True(Any(nil))
	a.True(MatcherFunc(Any).Match(nil))
}

func TestAnd(t *testing.T) {
	a := assert.New(t)

	hv := &HeaderVersion{Versions: []string{"1.0", "2.0"}}
	pv := NewPathVersion("v1", "v2")

	m := And(hv, pv)

	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.True(m.Match(r))

	// 未满足报头的值
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotNil(r)
	a.False(m.Match(r))
}

func TestOr(t *testing.T) {
	a := assert.New(t)

	hv := &HeaderVersion{Versions: []string{"1.0", "2.0"}}
	pv := NewPathVersion("v1", "v2")

	m := Or(hv, pv)

	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.True(m.Match(r))

	// 未满足报头的值，路径满足
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotNil(r)
	a.True(m.Match(r))

	// 所有条件都未满足
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/v111/test", nil)
	a.NotNil(r)
	a.False(m.Match(r))
}
