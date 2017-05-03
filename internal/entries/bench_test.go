// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entries

import (
	"testing"

	"github.com/issue9/assert"
)

// go1.8 BenchmarkCleanPath-4             	10000000	       144 ns/op
func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b)

	paths := []string{
		"/api//",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api/../",
		"/api/../../",
		"/api../",
	}

	for i := 0; i < b.N; i++ {
		ret := cleanPath(paths[i%len(paths)])
		a.True(len(ret) > 0)
	}
}
