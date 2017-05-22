// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"testing"

	"github.com/issue9/assert"
)

func TestByte_Add_Remove(t *testing.T) {
	a := assert.New(t)
	l := NewByte(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/1/author", h1))
	a.NotError(l.Add("/{posts}/1/*", h1))
	a.Equal(l.entries['p'].len(), 2)
	a.Equal(l.entries['{'].len(), 1)

	l.Remove("/posts/1")
	a.Equal(l.entries['p'].len(), 1)
	l.Remove("/{posts}/1/*")
	a.Equal(l.entries['{'].len(), 0)
}

func TestByte_Clean(t *testing.T) {
	a := assert.New(t)
	l := NewByte(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/1/author", h1))
	a.NotError(l.Add("/posts/1/*", h1))
	a.NotError(l.Add("/posts/tags/*", h1))
	a.NotError(l.Add("/posts/author", h1))

	l.Clean("/posts/1")
	a.Equal(l.entries['p'].len(), 2)

	l.Clean("")
	a.Equal(len(l.entries), 0)
}

func TestByte_Entry(t *testing.T) {
	a := assert.New(t)
	l := NewByte(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/tags/*", h1))

	a.Equal(l.entries['p'].len(), 2)
	e, err := l.Entry("/posts/tags/*")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/tags/*")

	// 不存在，自动添加
	e, err = l.Entry("/posts/1/author")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/1/author")
	a.Equal(l.entries['p'].len(), 3)
}

func TestByte_Match(t *testing.T) {
	a := assert.New(t)
	l := NewByte(false)
	a.NotNil(l)

	l.Add("/posts/{id}/*", h1) // 1
	l.Add("/posts/{id}/", h1)  // 2

	ety, ps := l.Match("/posts/1/")
	a.NotNil(ps).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/").
		Equal(ps, map[string]string{"id": "1"})

	ety, ps = l.Match("/posts/1/author")
	a.NotNil(ps).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/*").
		Equal(ps, map[string]string{"id": "1"})

	ety, ps = l.Match("/posts/1/author/profile")
	a.NotNil(ps).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/*").
		Equal(ps, map[string]string{"id": "1"})

	ety, ps = l.Match("/not-exists")
	a.Nil(ps).Nil(ety)
}

func TestByte_slashIndex(t *testing.T) {
	a := assert.New(t)
	l := &Byte{}
	a.Equal(l.slashIndex(countTestString), 'a')
	a.Equal(l.slashIndex("/{action}/1"), '{')
}
