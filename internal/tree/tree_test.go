// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import "net/http"

var (
	h1 = http.HandlerFunc(f1)
)

func f1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}
