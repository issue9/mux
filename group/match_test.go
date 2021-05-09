// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Matcher = MatcherFunc(Any)

func TestAny(t *testing.T) {
	a := assert.New(t)

	a.True(Any(nil))
	a.True(MatcherFunc(Any).Match(nil))
}
