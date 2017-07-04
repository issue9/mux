// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/issue9/mux/params"
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(nameStart), "(?P<",
	string(regexpSeparator), ">",
	string(nameEnd), ")")

type reg struct {
	name     string
	value    string
	endpoint bool
	expr     *regexp.Regexp
}

func newReg(str string) (Segment, error) {
	index := strings.IndexByte(str, regexpSeparator)

	r := repl.Replace(str)
	expr, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}

	return &reg{
		value:    str,
		name:     str[1:index],
		expr:     expr,
		endpoint: IsEndpoint(str),
	}, nil
}

func (r *reg) Type() Type {
	return TypeRegexp
}

func (r *reg) Value() string {
	return r.value
}

func (r *reg) Endpoint() bool {
	return r.endpoint
}

func (r *reg) Match(path string, params params.Params) (bool, string) {
	locs := r.expr.FindStringSubmatchIndex(path)
	if locs == nil || locs[0] != 0 { // 不匹配
		return false, path
	}

	params[r.name] = path[:locs[3]]
	return true, path[locs[1]:]
}

func (r *reg) DeleteParams(params params.Params) {
	delete(params, r.name)
}

func (r *reg) URL(buf *bytes.Buffer, params map[string]string) error {
	param, found := params[r.name]
	if !found {
		return fmt.Errorf("缺少参数 %s", r.name)
	}

	index := strings.IndexByte(r.value, nameEnd)
	url := strings.Replace(r.value, r.value[:index+1], param, 1)

	_, err := buf.WriteString(url)
	return err
}
