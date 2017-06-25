// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package method

import (
	"testing"

	"github.com/issue9/assert"
)

func TestIsSupported(t *testing.T) {
	a := assert.New(t)

	a.True(IsSupported("get"))
	a.True(IsSupported("POST"))
	a.False(IsSupported("not exists"))

	for method := range methodMap {
		a.True(IsSupported(method))
	}

	for _, method := range methodStringMap {
		a.True(IsSupported(method))
	}

	a.Equal(len(methodMap), len(methodStringMap))
}
