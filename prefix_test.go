// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix(t *testing.T) {
	a := assert.New(t)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewServeMux()
	p := mux.Prefix("p1")
	a.Equal(p.prefix, "p1").Equal(p.mux, mux)

	p.Post("/abc", hf)
	p.Get("/abc", hf)
	p.Delete("/abc", hf)

	a.Equal(1, mux.items["POST"].list.Len()).
		Equal(1, mux.items["GET"].list.Len()).
		Equal(1, mux.items["DELETE"].list.Len())
	p.Remove("/abc", "GET") // 从Prefix.Remove()删除
	a.Equal(1, mux.items["POST"].list.Len()).
		Equal(0, mux.items["GET"].list.Len()).
		Equal(1, mux.items["DELETE"].list.Len())
	mux.Remove("/abc", "POST") // 从ServeMux.Remove()删除，未带上p1前缀，无法删除
	a.Equal(1, mux.items["POST"].list.Len()).
		Equal(0, mux.items["GET"].list.Len()).
		Equal(1, mux.items["DELETE"].list.Len())
	mux.Remove("p1/abc", "POST") // 从ServeMux.Remove()删除
	a.Equal(0, mux.items["POST"].list.Len()).
		Equal(0, mux.items["GET"].list.Len()).
		Equal(1, mux.items["DELETE"].list.Len())
}
