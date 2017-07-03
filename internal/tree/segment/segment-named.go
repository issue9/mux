// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/issue9/mux/params"
)

type named struct {
	value    string
	name     string
	suffix   string
	endpoint bool
}

func newNamed(str string) (Segment, error) {
	endIndex := strings.IndexByte(str, nameEnd)
	if endIndex == -1 {
		return nil, fmt.Errorf("无效的路由语法：%s", str)
	}

	return &named{
		value:    str,
		endpoint: IsEndpoint(str),
		name:     str[1:endIndex],
		suffix:   str[endIndex+1:],
	}, nil
}

func (n *named) Type() Type {
	return TypeNamed
}

func (n *named) Value() string {
	return n.value
}

func (n *named) Endpoint() bool {
	return n.endpoint
}

func (n *named) Match(path string, params params.Params) (bool, string) {
	if n.endpoint {
		params[n.name] = path
		return true, path[:0]
	}

	index := strings.Index(path, n.suffix)
	if index > 0 { // 为零说明前面没有命名参数，肯定不正确
		params[n.name] = path[:index]
		return true, path[index+len(n.suffix):]
	}

	return false, path
}

func (n *named) DeleteParams(params params.Params) {
	delete(params, n.name)
}

func (n *named) Params(path string, params params.Params) string {
	if n.Endpoint() {
		params[n.name] = path
		return ""
	}

	index := strings.Index(path, n.suffix)
	params[n.name] = path[:index]
	return path[index+len(n.suffix):]
}

func (n *named) URL(buf *bytes.Buffer, params map[string]string) error {
	param, exists := params[n.name]
	if !exists {
		return fmt.Errorf("未找到参数 %s 的值", n.name)
	}
	buf.WriteString(param)
	buf.WriteString(n.suffix) // 如果是 endpoint suffix 肯定为空
	return nil
}
