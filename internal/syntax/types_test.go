// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

func TestGetType(t *testing.T) {
	a := assert.New(t)

	a.Equal(getType(""), String)
	a.Equal(getType("/posts"), String)
	a.Equal(getType("/posts/{id}"), Named)
	a.Equal(getType("/posts/{id}/author"), Named)
	a.Equal(getType("/posts/{id:\\d+}/author"), Regexp)
}

func TestType_String(t *testing.T) {
	a := assert.New(t)

	a.Equal(Named.String(), "named")
	a.Equal(Regexp.String(), "regexp")
	a.Equal(String.String(), "string")
	a.Panic(func() {
		_ = (Type(5)).String()
	})
}
