// SPDX-License-Identifier: MIT

// Package mux
package mux

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/internal/handlers"
)

func TestMethods(t *testing.T) {
	a := assert.New(t)
	a.Equal(Methods(), handlers.Methods)
}

func TestIsWell(t *testing.T) {
	a := assert.New(t)

	a.Error(IsWell("/{path"))
	a.Error(IsWell("/path}"))
	a.Error(IsWell(""))
}

func TestClearPath(t *testing.T) {
	a := assert.New(t)

	a.Equal(cleanPath(""), "/")

	a.Equal(cleanPath("/api//"), "/api/")
	a.Equal(cleanPath("api/"), "/api/")
	a.Equal(cleanPath("api/////"), "/api/")
	a.Equal(cleanPath("//api/////1"), "/api/1")

	a.Equal(cleanPath("/api/"), "/api/")
	a.Equal(cleanPath("/api/./"), "/api/./")

	a.Equal(cleanPath("/api/.."), "/api/..")
	a.Equal(cleanPath("/api/../"), "/api/../")
	a.Equal(cleanPath("/api/../../"), "/api/../../")
}
