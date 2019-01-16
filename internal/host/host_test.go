// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package host

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(false, true, false)
	a.False(h.disableOptions).
		False(h.skipCleanPath).
		True(h.disableHead).
		NotNil(h.tree)
}
