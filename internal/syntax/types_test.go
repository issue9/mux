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

	a.Equal(GetType(""), String)
	a.Equal(GetType("/posts"), String)
	a.Equal(GetType("/posts/{id}"), Named)
	a.Equal(GetType("/posts/{id}/author"), Named)
	a.Equal(GetType("/posts/{id:\\d+}/author"), Regexp)
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
