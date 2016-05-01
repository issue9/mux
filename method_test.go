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

func TestMethodsToInt(t *testing.T) {
	a := assert.New(t)

	a.Equal(get, methodsToInt("GET"))
	a.Equal(get, methodsToInt("GET", "GET"))
	a.Equal(get|post, methodsToInt("GET", "POST"))
	a.Equal(get|post, methodsToInt("GET", "POST", "POST"))
	a.Equal(get|post, methodsToInt("GET", "POST", "no exists"))
}

func TestGetAllowString(t *testing.T) {
	a := assert.New(t)

	a.Equal("GET", getAllowString(get))
	a.Equal("GET POST", getAllowString(get|post))
	a.Equal("GET POST", getAllowString(get|post|post))
	a.Equal("GET OPTIONS POST", getAllowString(get|post|options))
}
