// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"strings"

	"github.com/issue9/mux/params"
)

type str string

func (s str) Type() Type {
	return TypeString
}

func (s str) Value() string {
	return string(s)
}

func (s str) Endpoint() bool {
	return false
}

func (s str) Match(path string) (bool, string) {
	if strings.HasPrefix(path, string(s)) {
		return true, path[len(s):]
	}

	return false, path
}

func (s str) Params(path string, params params.Params) string {
	return path[len(s):]
}

func (s str) URL(buf *bytes.Buffer, params map[string]string) error {
	buf.WriteString(string(s))
	return nil
}
