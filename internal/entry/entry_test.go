// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

type matcher struct {
	a *assert.Assertion
	e Entry
}

func newMatcher(a *assert.Assertion, pattern string) *matcher {
	e, err := New(pattern, nil)
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

func TestNew(t *testing.T) {
	//a := assert.New(t)
	//  TODO
}
