// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestInStringSlice(t *testing.T) {
	a := assert.New(t)

	slice := []string{"a", "b", "c"}
	a.True(inStringSlice(slice, "a"))
	a.True(inStringSlice(slice, "b"))
	a.False(inStringSlice(slice, "C"))
}
