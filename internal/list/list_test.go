// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"testing"

	"github.com/issue9/assert"
)

const countTestString = "/adfada/adfa/dd//adfadasd/ada/dfad/"

func TestList_entriesIndex(t *testing.T) {
	a := assert.New(t)
	l := &List{}
	a.Equal(l.entriesIndex(countTestString), 8)
	a.Equal(l.entriesIndex(countTestString+"*"), wildcardEntriesIndex)
}
