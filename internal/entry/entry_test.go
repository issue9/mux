// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

// 用于测试 Entry.match 的对象
type matcher struct {
	a *assert.Assertion
	e Entry
}

func newMatcher(a *assert.Assertion, pattern string) *matcher {
	e, err := NewEntry(pattern, nil)
	a.NotError(err).NotNil(e)

	return &matcher{
		a: a,
		e: e,
	}
}

func (m *matcher) True(path string, params map[string]string) *matcher {
	ok, ps := m.e.match(path)
	m.a.True(ok).
		Equal(ps, params)

	return m
}

func (m *matcher) False(path string, params map[string]string) *matcher {
	ok, _ := m.e.match(path) // 为 false 则返回值没意义
	m.a.False(ok)

	return m
}

func TestNewEntry(t *testing.T) {
	a := assert.New(t)

	// basic
	e, err := NewEntry("/basic/basic", nil)
	a.NotError(err).NotNil(e)
	b, ok := e.(*basic)
	a.True(ok).False(b.wildcard)

	// basic with wildcard
	e, err = NewEntry("/basic/basic/*", nil)
	a.NotError(err).NotNil(e)
	b, ok = e.(*basic)
	a.True(ok).True(b.wildcard)

	// named
	e, err = NewEntry("/named/{named}/path", nil)
	a.NotError(err).NotNil(e)
	n, ok := e.(*named)
	a.True(ok).False(n.wildcard)

	// named with wildcard
	e, err = NewEntry("/named/{named}/path/*", nil)
	a.NotError(err).NotNil(e)
	n, ok = e.(*named)
	a.True(ok).True(n.wildcard)

	// regexp
	e, err = NewEntry("/regexp/{named:\\d+}", nil)
	a.NotError(err).NotNil(e)
	r, ok := e.(*regexp)
	a.True(ok).False(r.wildcard)

	// regexp with wildcard
	e, err = NewEntry("/regexp/{named:\\d+}/*", nil)
	a.NotError(err).NotNil(e)
	r, ok = e.(*regexp)
	a.True(ok).True(r.wildcard)
}

func TestEntry_priority(t *testing.T) {
	a := assert.New(t)

	b, err := NewEntry("/basic/basic", nil)
	a.NotError(err).NotNil(b)

	bw, err := NewEntry("/basic/basic/*", nil)
	a.NotError(err).NotNil(bw)

	n, err := NewEntry("/basic/{named}", nil)
	a.NotError(err).NotNil(n)

	nw, err := NewEntry("/basic/{named}/*", nil)
	a.NotError(err).NotNil(nw)

	r, err := NewEntry("/basic/{named:\\d+}", nil)
	a.NotError(err).NotNil(r)

	rw, err := NewEntry("/basic/{named:\\d+}/*", nil)
	a.NotError(err).NotNil(rw)

	a.True(bw.priority() > b.priority()).
		True(nw.priority() > n.priority()).
		True(rw.priority() > r.priority())

	a.True(n.priority() > r.priority()).
		True(r.priority() > b.priority())

	a.True(nw.priority() > rw.priority()).
		True(rw.priority() > bw.priority())

	a.True(bw.priority() > n.priority())
}
