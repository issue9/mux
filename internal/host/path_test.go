// SPDX-License-Identifier: MIT

package host

import (
	"testing"

	"github.com/issue9/assert"
)

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

func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b)

	paths := []string{
		"",
		"/api//",
		"/api////users/1",
		"//api/users/1",
		"api///users////1",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api//../",
		"/api/..//../",
		"/api../",
		"api../",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ret := cleanPath(paths[i%len(paths)])
		a.True(len(ret) > 0)
	}
}
