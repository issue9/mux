// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import "net/http"

var (
	h1 = http.HandlerFunc(f1)
	h2 = http.HandlerFunc(f2)
	h3 = http.HandlerFunc(f3)
	h4 = http.HandlerFunc(f4)
)

func f1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

func f2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(2)
}

func f3(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(3)
}

func f4(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(4)
}
