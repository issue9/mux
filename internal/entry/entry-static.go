// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

// 静态文件匹配路由项，只要路径中的开头字符串与 pattern 相同，
// 且 pattern 以 / 结尾，即表示匹配成功。根据 match() 的返回值来确定哪个最匹配。
type static struct {
	*base
}

func (s *static) Type() int {
	return TypeStatic
}

func (s *static) Params(url string) map[string]string {
	return nil
}

func (s *static) Match(url string) int {
	l := len(url) - len(s.pattern)
	if l < 0 {
		return -1
	}

	// 由 New 函数确保 s.pattern 都是以 '/' 结尾的
	if s.pattern == url[:len(s.pattern)] {
		return l
	}
	return -1
}

func (s *static) URL(params map[string]string) (string, error) {
	return s.pattern, nil
}
