// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"strings"

	"github.com/issue9/mux/internal/syntax"
)

// 最基本的字符串匹配，只能全字符串匹配。
type basic struct {
	*base
	prefix string
}

func newBasic(s *syntax.Syntax) *basic {
	ret := &basic{
		base: newBase(s.Pattern),
	}
	if ret.wildcard {
		ret.prefix = s.Pattern[:len(s.Pattern)-1]
	}

	return ret
}

func (b *basic) Priority() int {
	return syntax.TypeBasic
}

func (b *basic) Match(url string) (bool, map[string]string) {
	if b.wildcard {
		return strings.HasPrefix(url, b.prefix), nil
	}

	return url == b.pattern, nil
}

func (b *basic) URL(params map[string]string, path string) (string, error) {
	if b.wildcard {
		return b.prefix + path, nil
	}

	return b.Pattern(), nil
}
