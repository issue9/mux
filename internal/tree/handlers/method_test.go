// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMethods(t *testing.T) {
	a := assert.New(t)

	for typ, method := range methodMap {
		a.Equal(typ, methodStringMap[method])
	}

	a.Equal(len(methodMap), len(methodStringMap))
}
