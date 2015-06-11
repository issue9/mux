// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSplit(t *testing.T) {
	a := assert.New(t)

	a.Equal(split("/blog/post/1"), []string{"/blog/post/1"})
	a.Equal(split("/blog/post/{id}"), []string{"/blog/post/", "{id}"})
	a.Equal(split("/blog/post/{id:\\d}"), []string{"/blog/post/", "{id:\\d}"})
	a.Equal(split("/blog/{post}/{id}"), []string{"/blog/", "{post}", "/", "{id}"})
	a.Equal(split("/blog/{post}-{id}"), []string{"/blog/", "{post}", "-", "{id}"})

	a.Equal(split("{/blog/post/{id}"), []string{"{/blog/post/{id}"})
	a.Equal(split("}/blog/post/{id}"), []string{"}/blog/post/", "{id}"})
}

func TestNewEntry(t *testing.T) {
	a := assert.New(t)

	a.Equal(newEntry("/blog/post/1"), staticEntry("/blog/post/1"))
	//a.Equal(newEntry("/blog/{post}/1"), ("/blog/{post/1"))
}
