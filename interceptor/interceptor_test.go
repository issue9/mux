// SPDX-License-Identifier: MIT

package interceptor

import (
	"testing"

	"github.com/issue9/assert"
)

func TestInterceptors(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		Register(MatchAny)
	})

	a.NotError(Register(MatchWord, "[a-zA-Z0-9]+", "word1", "word2"))
	a.Error(Register(MatchWord, "[a-zA-Z0-9]+")) // 已经存在
	_, found := Get("word1")
	a.True(found)
	_, found = Get("[a-zA-Z0-9]+")
	a.True(found)

	Deregister("word1", "word2")
	_, found = Get("word1")
	a.False(found)
	_, found = Get("[a-zA-Z0-9]+")
	a.True(found)
}
