// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

var _ error = &Error{}

func TestGetType(t *testing.T) {
	a := assert.New(t)

	a.Equal(GetType(""), TypeString)
	a.Equal(GetType("/posts"), TypeString)
	a.Equal(GetType("/posts/{id}"), TypeNamed)
	a.Equal(GetType("/posts/{id}/author"), TypeNamed)
	a.Equal(GetType("/posts/{id:\\d+}/author"), TypeRegexp)
}
