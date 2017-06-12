// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestTree_Trace(t *testing.T) {
	a := assert.New(t)
	tree := New()

	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(1), http.MethodGet)) //

	a.Equal(tree.Trace("/"), []string{"", "/"})
	a.Equal(tree.Trace("/posts/1.html/page"), []string{"", "/", "posts/", "{id}"})
}
