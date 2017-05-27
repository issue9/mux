// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

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

func TestCheck(t *testing.T) {
	a := assert.New(t)

	test := func(pattern string, nType Type, isError bool) {
		typ, err := Check(pattern)
		if isError {
			a.Error(err)
			return
		}

		a.NotError(err).Equal(nType, typ)
	}

	test("{{", TypeUnknown, true)
	test("}{", TypeUnknown, true)
	test("{}:", TypeUnknown, true)
	test("{}", TypeUnknown, true)
	test("{:}", TypeUnknown, true)
	test("{a:}:", TypeUnknown, true)

	test("{a:}", TypeRegexp, false)
	test("{a:\\d+}", TypeRegexp, false)
	test("{a}", TypeNamed, false)
	test("{a}/*", TypeWildcard, false)
	test("{a:\\d+}/*", TypeWildcard, false)
}
