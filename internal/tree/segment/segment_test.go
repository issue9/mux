// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
)

func TestStringType(t *testing.T) {
	a := assert.New(t)

	a.Equal(stringType("/posts"), TypeString)
	a.Equal(stringType("/posts/{id}"), TypeNamed)
	a.Equal(stringType("/posts/{id}/author"), TypeNamed)
	a.Equal(stringType("/posts/{id:\\d+}/author"), TypeRegexp)
}

func TestEqaul(t *testing.T) {
	a := assert.New(t)

	s1 := str("")
	s2 := &named{}
	a.False(Equal(s1, s2))
}
