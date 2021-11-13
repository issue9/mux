// SPDX-License-Identifier: MIT

package options

import (
	"testing"

	"github.com/issue9/assert"
)

func TestURL_sanitize(t *testing.T) {
	a := assert.New(t)

	u := &URL{}
	u.sanitize()
	a.Empty(u.Domain)

	u = &URL{
		Domain: "https://example.com",
	}
	u.sanitize()
	a.Equal(u.Domain, "https://example.com")

	u = &URL{
		Domain: "https://example.com/",
	}
	u.sanitize()
	a.Equal(u.Domain, "https://example.com")
}
