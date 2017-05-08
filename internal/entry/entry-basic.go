// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"strings"
)

// 最基本的字符串匹配，只能全字符串匹配。
type basic struct {
	*base
	prefix string
}

func newBasic(pattern string) *basic {
	ret := &basic{
		base: newBase(pattern),
	}
	if ret.wildcard {
		ret.prefix = pattern[:len(pattern)-1]
	}

	return ret
}

func (b *basic) priority() int {
	if b.wildcard {
		return typeBasic + 100
	}

	return typeBasic
}

func (b *basic) Params(url string) map[string]string {
	return nil
}

func (b *basic) match(url string) bool {
	if b.wildcard {
		return strings.HasPrefix(url, b.prefix)
	}

	return url == b.pattern
}

func (b *basic) URL(params map[string]string) (string, error) {
	return b.pattern, nil
}
