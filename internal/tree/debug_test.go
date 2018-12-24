// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestTree_Trace(t *testing.T) {
	a := assert.New(t)
	tree := New(false, false)

	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author/profile", buildHandler(1), http.MethodGet))

	//tree.Trace(os.Stdout, "/")
	tree.Trace(os.Stdout, "/posts/1.html/author/profile")
}
