// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/syntax"
)

// 用于测试 Entry.match 的对象
type matcher struct {
	a *assert.Assertion
	e Entry
}

func newMatcher(a *assert.Assertion, pattern string) *matcher {
	s, err := syntax.New(pattern)
	a.NotError(err).NotNil(s)
	e, err := New(s)
	a.NotError(err).NotNil(e)

	return &matcher{
		a: a,
		e: e,
	}
}

func (m *matcher) True(path string, params map[string]string) *matcher {
	ok, ps := m.e.Match(path)
	m.a.True(ok).
		Equal(ps, params)

	return m
}

func (m *matcher) False(path string, params map[string]string) *matcher {
	ok, _ := m.e.Match(path) // 为 false 则返回值没意义
	m.a.False(ok)

	return m
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	// basic
	s, err := syntax.New("/basic/basic")
	a.NotError(err).NotNil(s)
	e, err := New(s)
	a.NotError(err).NotNil(e)
	b, ok := e.(*basic)
	a.True(ok).False(b.wildcard)

	// basic with wildcard
	s, err = syntax.New("/basic/basic/*")
	a.NotError(err).NotNil(s)
	e, err = New(s)
	a.NotError(err).NotNil(e)
	b, ok = e.(*basic)
	a.True(ok).True(b.wildcard)

	// named
	s, err = syntax.New("/named/{named}/path")
	a.NotError(err).NotNil(s)
	e, err = New(s)
	a.NotError(err).NotNil(e)
	n, ok := e.(*named)
	a.True(ok).False(n.wildcard)

	// named with wildcard
	s, err = syntax.New("/named/{named}/path/*")
	a.NotError(err).NotNil(s)
	e, err = New(s)
	a.NotError(err).NotNil(e)
	n, ok = e.(*named)
	a.True(ok).True(n.wildcard)

	// regexp
	s, err = syntax.New("/regexp/{named:\\d+}")
	a.NotError(err).NotNil(s)
	e, err = New(s)
	a.NotError(err).NotNil(e)
	r, ok := e.(*regexp)
	a.True(ok).False(r.wildcard)

	// regexp with wildcard
	s, err = syntax.New("/regexp/{named:\\d+}/*")
	a.NotError(err).NotNil(s)
	e, err = New(s)
	a.NotError(err).NotNil(e)
	r, ok = e.(*regexp)
	a.True(ok).True(r.wildcard)
}

func TestEntry_Priority(t *testing.T) {
	a := assert.New(t)

	s, err := syntax.New("/basic/basic")
	a.NotError(err).NotNil(s)
	b, err := New(s)
	a.NotError(err).NotNil(b)

	s, err = syntax.New("/basic/basic/*")
	a.NotError(err).NotNil(s)
	bw, err := New(s)
	a.NotError(err).NotNil(bw)

	s, err = syntax.New("/basic/{named}")
	a.NotError(err).NotNil(s)
	n, err := New(s)
	a.NotError(err).NotNil(n)

	s, err = syntax.New("/basic/{named}/*")
	a.NotError(err).NotNil(s)
	nw, err := New(s)
	a.NotError(err).NotNil(nw)

	s, err = syntax.New("/basic/{named:\\d+}")
	a.NotError(err).NotNil(s)
	r, err := New(s)
	a.NotError(err).NotNil(r)

	s, err = syntax.New("/basic/{named:\\d+}/*")
	a.NotError(err).NotNil(s)
	rw, err := New(s)
	a.NotError(err).NotNil(rw)

	a.True(n.Priority() > r.Priority()).
		True(r.Priority() > b.Priority())
}
