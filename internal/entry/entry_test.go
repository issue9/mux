// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/syntax"
)

func newEntry(pattern string) (Entry, error) {
	s, err := syntax.New(pattern)
	if err != nil {
		return nil, err
	}

	return New(s)
}

// 用于测试 Entry.match 的对象
type matcher struct {
	a *assert.Assertion
	e Entry
}

func newMatcher(a *assert.Assertion, pattern string) *matcher {
	e, err := newEntry(pattern)
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
	e, err := newEntry("/basic/basic")
	a.NotError(err).NotNil(e)
	b, ok := e.(*basic)
	a.True(ok).False(b.wildcard)

	// basic with wildcard
	e, err = newEntry("/basic/basic/*")
	a.NotError(err).NotNil(e)
	b, ok = e.(*basic)
	a.True(ok).True(b.wildcard)

	// named
	e, err = newEntry("/named/{named}/path")
	a.NotError(err).NotNil(e)
	n, ok := e.(*named)
	a.True(ok).False(n.wildcard)

	// named with wildcard
	e, err = newEntry("/named/{named}/path/*")
	a.NotError(err).NotNil(e)
	n, ok = e.(*named)
	a.True(ok).True(n.wildcard)

	// regexp
	e, err = newEntry("/regexp/{named:\\d+}")
	a.NotError(err).NotNil(e)
	r, ok := e.(*regexp)
	a.True(ok).False(r.wildcard)

	// regexp with wildcard
	e, err = newEntry("/regexp/{named:\\d+}/*")
	a.NotError(err).NotNil(e)
	r, ok = e.(*regexp)
	a.True(ok).True(r.wildcard)
}

func TestEntry_Priority(t *testing.T) {
	a := assert.New(t)

	b, err := newEntry("/basic/basic")
	a.NotError(err).NotNil(b)

	bw, err := newEntry("/basic/basic/*")
	a.NotError(err).NotNil(bw)

	n, err := newEntry("/basic/{named}")
	a.NotError(err).NotNil(n)

	nw, err := newEntry("/basic/{named}/*")
	a.NotError(err).NotNil(nw)

	r, err := newEntry("/basic/{named:\\d+}")
	a.NotError(err).NotNil(r)

	rw, err := newEntry("/basic/{named:\\d+}/*")
	a.NotError(err).NotNil(rw)

	a.True(r.Priority() > n.Priority()).
		True(n.Priority() > b.Priority())

	a.True(rw.Priority() > r.Priority()).
		True(nw.Priority() > n.Priority()).
		True(bw.Priority() > b.Priority())

	a.True(bw.Priority() > r.Priority())
}
