// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestStatusError(t *testing.T) {
	err := &StatusError{Code: 404, Message: "not found"}
	assert.Equal(t, err.Error(), "404:not found")
}
