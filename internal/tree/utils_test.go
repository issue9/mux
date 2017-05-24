// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"testing"

	"github.com/issue9/assert"
)

func TestPrefixLen(t *testing.T) {
	a := assert.New(t)

	a.Equal(prefixLen("", ""), 0)
	a.Equal(prefixLen("/", ""), 0)
	a.Equal(prefixLen("/test", "test"), 0)
	a.Equal(prefixLen("/test", "/abc"), 1)
	a.Equal(prefixLen("/test", "/test"), 5)
	a.Equal(prefixLen("/te{st", "/test"), 3)
	a.Equal(prefixLen("/test", "/tes{t"), 4)
	a.Equal(prefixLen("/tes{t}", "/tes{t}"), 7)
}

func TestCheckSyntax(t *testing.T) {
	a := assert.New(t)

	a.Error(checkSyntax("{{"))
	a.Error(checkSyntax("}{"))

	a.NotError(checkSyntax("{}"))
	a.NotError(checkSyntax(":{}"))
}
