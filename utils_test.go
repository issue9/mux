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

func TestDeleteStringsSlice(t *testing.T) {
	a := assert.New(t)

	s1 := []string{"1", "2", "3333", "4", "567"}
	s2 := []string{"1", "2", "567", "789"}
	slice := deleteStringsSlice(s1, s2...)
	a.Equal(slice, []string{"3333", "4"})
}
