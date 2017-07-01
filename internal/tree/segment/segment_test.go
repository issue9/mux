// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Segment = str("")
var _ Segment = &named{}
var _ Segment = &reg{}

func TestStringType(t *testing.T) {
	a := assert.New(t)

	a.Equal(stringType("/posts"), TypeString)
	a.Equal(stringType("/posts/{id}"), TypeNamed)
	a.Equal(stringType("/posts/{id}/author"), TypeNamed)
	a.Equal(stringType("/posts/{id:\\d+}/author"), TypeRegexp)
}

func TestEqaul(t *testing.T) {
	a := assert.New(t)

	a.False(Equal(str(""), &named{}))
	a.True(Equal(&named{}, &named{}))

	s1, err := newNamed("{action}")
	a.NotError(err).NotNil(s1)
	s2, err := newNamed("{action}")
	a.NotError(err).NotNil(s2)
	a.True(Equal(s1, s2))

	s2, err = newNamed("{action}/1")
	a.NotError(err).NotNil(s2)
	a.False(Equal(s1, s2))
}
