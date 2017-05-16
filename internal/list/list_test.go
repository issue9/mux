// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

const countTestString = "/adfada/adfa/dd//adfadasd/ada/dfad/"

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	h1 = http.HandlerFunc(f1)
)

func TestList_Add_Remove(t *testing.T) {
	a := assert.New(t)
	l := New(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/1/author", h1))
	a.NotError(l.Add("/posts/1/*", h1))
	a.Equal(l.entries[2].len(), 1)
	a.Equal(l.entries[3].len(), 1)
	a.Equal(l.entries[wildcardEntriesIndex].len(), 1)

	l.Remove("/posts/1")
	a.Equal(l.entries[2].len(), 0)
	l.Remove("/posts/1/*")
	a.Equal(l.entries[wildcardEntriesIndex].len(), 0)
}

func TestList_Clean(t *testing.T) {
	a := assert.New(t)
	l := New(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/1/author", h1))
	a.NotError(l.Add("/posts/1/*", h1))
	a.NotError(l.Add("/posts/tags/*", h1))
	a.NotError(l.Add("/posts/author", h1))

	l.Clean("/posts/1")
	a.Equal(l.entries[2].len(), 1) // 还有 /posts/author
	a.Equal(l.entries[3].len(), 0)
	a.Equal(l.entries[wildcardEntriesIndex].len(), 1) // 还有 /posts/tags/*

	l.Clean("")
	a.Equal(len(l.entries), 0)
}

func TestList_Entry(t *testing.T) {
	a := assert.New(t)
	l := New(false)

	a.NotError(l.Add("/posts/1", h1))
	a.NotError(l.Add("/posts/tags/*", h1))

	a.Equal(l.entries[2].len(), 1)
	e, err := l.Entry("/posts/tags/*")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/tags/*")
	a.Equal(l.entries[2].len(), 1)

	// 不存在，自动添加
	a.Nil(l.entries[3])
	e, err = l.Entry("/posts/1/author")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/1/author")
	a.Equal(l.entries[3].len(), 1)
}

func TestList_Match(t *testing.T) {
	a := assert.New(t)
	l := New(false)
	a.NotNil(l)

	l.Add("/posts/{id}/*", h1) // 1
	l.Add("/posts/{id}/", h1)  // 2

	ety, err := l.Match("/posts/1/")
	a.NotError(err).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/")

	ety, err = l.Match("/posts/1/author")
	a.NotError(err).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/*")

	ety, err = l.Match("/posts/1/author/profile")
	a.NotError(err).NotNil(ety)
	a.Equal(ety.Pattern(), "/posts/{id}/*")
}

func TestList_entriesIndex(t *testing.T) {
	a := assert.New(t)
	l := &List{}
	a.Equal(l.entriesIndex(countTestString), 8)
	a.Equal(l.entriesIndex(countTestString+"*"), wildcardEntriesIndex)
}
