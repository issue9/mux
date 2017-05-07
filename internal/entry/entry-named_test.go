// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNamed_Type(t *testing.T) {
	a := assert.New(t)
	n := &named{}
	a.Equal(n.Type(), TypeNamed)
}
