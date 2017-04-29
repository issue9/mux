// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMethodIsSupported(t *testing.T) {
	a := assert.New(t)

	a.True(MethodIsSupported("get"))
	a.True(MethodIsSupported("POST"))
	a.False(MethodIsSupported("not exists"))

	// defaultMethods 必然属于支付列表中的一员
	for _, method := range defaultMethods {
		a.True(MethodIsSupported(method))
	}
}
