// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"strings"
)

type str struct {
	pattern string
}

func newStr(s string) (Segment, error) {
	return &str{
		pattern: s,
	}, nil
}

func (s *str) Type() Type {
	return TypeString
}

func (s *str) Pattern() string {
	return s.pattern
}

func (s *str) Endpoint() bool {
	return false
}

func (s *str) Match(path string) (bool, string) {
	if strings.HasPrefix(path, s.pattern) {
		return true, path[len(s.pattern):]
	}

	return false, path
}

func (s *str) Params(path string, params map[string]string) string {
	return path[len(s.pattern):]
}

func (s *str) URL(buf *bytes.Buffer, params map[string]string) error {
	buf.WriteString(s.pattern)
	return nil
}
