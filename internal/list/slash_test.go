// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"net/http"
	"strings"
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

func TestSlash_add_Remove(t *testing.T) {
	a := assert.New(t)
	l := newSlash(false)

	a.NotError(l.add(newSyntax(a, "/posts/1/detail"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/1/author"), h1))
	a.NotError(l.add(newSyntax(a, "/{posts}/1/*"), h1))
	a.Equal(l.entries[3].len(), 2)
	a.Equal(l.entries[wildcardIndex].len(), 1)

	l.Remove("/posts/1/detail")
	a.Equal(l.entries[3].len(), 1)
	l.Remove("/{posts}/1/*")
	a.Equal(l.entries[wildcardIndex].len(), 0)
}

func TestSlash_Clean(t *testing.T) {
	a := assert.New(t)
	l := newSlash(false)

	a.NotError(l.add(newSyntax(a, "/posts/1"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/1/author"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/1/*"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/tags/*"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/author"), h1))

	l.Clean("/posts/1")
	a.Equal(l.entries[2].len(), 1)

	l.Clean("")
	a.Equal(len(l.entries), 0)
}

func TestSlash_Entry(t *testing.T) {
	a := assert.New(t)
	l := newSlash(false)

	a.NotError(l.add(newSyntax(a, "/posts/1"), h1))
	a.NotError(l.add(newSyntax(a, "/posts/tags/*"), h1))

	a.Equal(l.entries[2].len(), 1)
	a.Equal(l.entries[wildcardIndex].len(), 1)
	e, err := l.Entry("/posts/tags/*")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/tags/*")

	// 不存在，自动添加
	e, err = l.Entry("/posts/1/author")
	a.NotError(err).NotNil(e)
	a.Equal(e.Pattern(), "/posts/1/author")
	a.Equal(l.entries[3].len(), 1)
}

func TestSlash_Match(t *testing.T) {
	a := assert.New(t)
	l := newSlash(false)
	a.NotNil(l)

	a.NotError(l.add(newSyntax(a, "/posts/{id}/*"), h1)) // 1
	a.NotError(l.add(newSyntax(a, "/posts/{id}/"), h1))  // 2

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

func TestSlash_entriesIndex(t *testing.T) {
	a := assert.New(t)
	l := &slash{}

	a.Equal(l.entriesIndex(newSyntax(a, countTestString)), 8)
	a.Equal(l.entriesIndex(newSyntax(a, "/{action}/1")), 2)
}

func TestByteCount(t *testing.T) {
	a := assert.New(t)
	a.Equal(byteCount('/', countTestString), 8)
}

func BenchmarkStringsCount(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if strings.Count(countTestString, "/") != 8 {
			b.Error("strings.Count:error")
		}
	}
}

func BenchmarkSlashCount(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if byteCount('/', countTestString) != 8 {
			b.Error("count:error")
		}
	}
}
