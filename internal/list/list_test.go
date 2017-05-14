// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"strings"
	"testing"

	"github.com/issue9/assert"
)

const countTestString = "/adfada/adfa/dd//adfadasd/ada/dfad/"

func TestGetSlashSize(t *testing.T) {
	a := assert.New(t)
	a.Equal(getSlashSize(countTestString), 8)
	a.Equal(getSlashSize(countTestString+"*"), maxSlashSize)
}

func TestSlashCount(t *testing.T) {
	a := assert.New(t)
	a.Equal(slashCount(countTestString), 8)
}

func BenchmarkStringsCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if strings.Count(countTestString, "/") != 8 {
			b.Error("strings.count.error")
		}
	}
}

func BenchmarkSlashCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if slashCount(countTestString) != 8 {
			b.Error("count:error")
		}
	}
}
