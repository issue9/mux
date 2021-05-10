// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Matcher = MatcherFunc(Any)

func TestAny(t *testing.T) {
	a := assert.New(t)

	r, ok := Any(nil)
	a.True(ok).Nil(r)

	r, ok = MatcherFunc(Any).Match(nil)
	a.True(ok).Nil(r)
}
