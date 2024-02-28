// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package options

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o, err := Build()
	a.NotError(err).
		NotNil(o).
		NotNil(o.CORS)

	// URLDomain

	o, err = Build(func(o *Options) { o.URLDomain = "https://example.com" })
	a.NotError(err).NotNil(o).Equal(o.URLDomain, "https://example.com")

	o, err = Build(func(o *Options) { o.URLDomain = "https://example.com/" })
	a.NotError(err).NotNil(o).Equal(o.URLDomain, "https://example.com")

	o, err = Build(func(o *Options) { o.CORS = &CORS{AllowCredentials: true, Origins: []string{"*"}} })
	a.Error(err).Nil(o)
}
