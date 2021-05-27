// SPDX-License-Identifier: MIT

package mux

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/internal/tree"
)

func TestMethods(t *testing.T) {
	a := assert.New(t)
	a.Equal(Methods(), tree.Methods)
}

func TestCheckSyntax(t *testing.T) {
	a := assert.New(t)

	a.NotError(CheckSyntax("/{path"))
	a.NotError(CheckSyntax("/path}"))
	a.Error(CheckSyntax(""))
}
