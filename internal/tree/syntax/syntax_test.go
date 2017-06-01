// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParse(t *testing.T) {
	a := assert.New(t)
	test := func(str string, isError bool, ss ...*Segment) {
		s, err := Parse(str)
		if isError {
			a.Error(err)
			return
		}

		a.NotError(err).
			Equal(len(s), len(ss))
		for index, seg := range ss {
			a.Equal(seg, s[index])
		}
	}

	test("/", false, &Segment{Value: "/", Type: TypeBasic})
	test("/posts/1", false, &Segment{Value: "/posts/1", Type: TypeBasic})

	test("/posts/{id}", false, &Segment{Value: "/posts/", Type: TypeBasic},
		&Segment{Value: "{id}", Type: TypeNamed})
	test("/posts/{id}/author/profile", false, &Segment{Value: "/posts/", Type: TypeBasic},
		&Segment{Value: "{id}/author/profile", Type: TypeNamed})

	test("/posts/{id}/page/{page}", false, &Segment{Value: "/posts/", Type: TypeBasic},
		&Segment{Value: "{id}/page/", Type: TypeNamed},
		&Segment{Value: "{page}", Type: TypeNamed})

	// 正则
	test("/posts/{id:\\d+}", false, &Segment{Value: "/posts/", Type: TypeBasic},
		&Segment{Value: "{id:\\d+}", Type: TypeRegexp})

	test("/posts/{id:\\d+}/*", false, &Segment{Value: "/posts/", Type: TypeBasic},
		&Segment{Value: "{id:\\d+}/*", Type: TypeWildcard})
}

func TestPrefixLen(t *testing.T) {
	a := assert.New(t)

	a.Equal(PrefixLen("", ""), 0)
	a.Equal(PrefixLen("/", ""), 0)
	a.Equal(PrefixLen("/test", "test"), 0)
	a.Equal(PrefixLen("/test", "/abc"), 1)
	a.Equal(PrefixLen("/test", "/test"), 5)
	a.Equal(PrefixLen("/te{st", "/test"), 3)
	a.Equal(PrefixLen("/test", "/tes{t"), 4)
	a.Equal(PrefixLen("/tes{t}", "/tes{t}"), 7)
	a.Equal(PrefixLen("/tes{t:\\d+}", "/tes{t:\\d+}"), 4)
}
