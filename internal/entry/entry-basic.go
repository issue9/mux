// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import "strings"

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
		return typeBasicWithWildcard
	}

	return typeBasic
}

func (b *basic) match(url string) (bool, map[string]string) {
	if b.wildcard {
		return strings.HasPrefix(url, b.prefix), nil
	}

	return url == b.patternString, nil
}

func (b *basic) URL(params map[string]string, path string) (string, error) {
	if b.wildcard {
		return b.prefix + path, nil
	}

	return b.pattern(), nil
}
