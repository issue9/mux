// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"strings"
	"testing"
)

const tt = "/adfada/adfa/dd//adfadasd/ada/dfad/"

func BenchmarkStringsCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if strings.Count(tt, "/") != 8 {
			b.Error("strings.count.error")
		}
	}
}

func BenchmarkSlashCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if slashCount(tt) != 8 {
			b.Error("count:error")
		}
	}
}
