// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package syntax

import "fmt"

// InterceptorFunc 拦截器的处理函数
type InterceptorFunc func(string) bool

type Interceptors struct {
	funcs map[string]InterceptorFunc
}

func NewInterceptors() *Interceptors {
	return &Interceptors{
		funcs: map[string]InterceptorFunc{},
	}
}

func (i *Interceptors) Add(f InterceptorFunc, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		if _, found := i.funcs[n]; found {
			panic(fmt.Sprintf("%s 已经存在", n))
		}
		i.funcs[n] = f
	}
}

// MatchAny 匹配任意非空内容
func MatchAny(path string) bool { return len(path) > 0 }

// MatchDigit 匹配数值字符
//
// 与正则表达式中的 [0-9]+ 是相同的。
func MatchDigit(path string) bool {
	for _, c := range path {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(path) > 0
}

// MatchWord 匹配单词
//
// 与正则表达式中的 [a-zA-Z0-9]+ 是相同的。
func MatchWord(path string) bool {
	for _, c := range path {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return len(path) > 0
}
