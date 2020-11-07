// SPDX-License-Identifier: MIT

// Package interceptor 针对正则表达式拦截处理
//
// 在解析诸如 /authors/{id:\\d+} 带正则表达式的路由项时，
// 用户可以通过拦截并自定义对 {id:\\d+} 的解析，
// 从而不需要走正则表达式的那一套解析流程，可以在一定程度上增强性能。
//
// 一旦正则表达式被拦截，则节点类型也将不再是正则表达式，
// 而是命名参数，其处理优先级会比正则表达式类型的低。
// 在某些情况下，可能会造成处理结果不相同。比如：
//  /authors/{id:\\d+}     // 1
//  /authors/{id:[0-9]+}   // 2
// 以上两条记录是相同的，但因为表达式不同，也能正常添加，
// 处理流程，会按添加顺序优先比对第一条，所以第二条是永远无法匹配的。
// 但是如果你此时添加了 Register(MatchDigit, "\\d+")，
// 将第一个记录的类型降为了命名节点，以后的匹配都是优先第二条，
// 造成第一条永远无法匹配到数据。
//
// 除非是改造旧有的项目，否则建议自定义一些约束符来处理。比如：
//  /authors/{id:digit}
// 用户只要注册一个 Register(MatchDigit, "digit") 即可拦截针对 digit 的项，
// 且不会影响正则表达式的处理。
//
// interceptor 也是本着这样的原则，添加了以下拦截器：
//  - digit 数字；
//  - word 单词，即 [a-zA-Z0-9]+；
package interceptor

import (
	"fmt"
	"sync"
)

var interceptors = &sync.Map{}

// Register 注册拦截器
//
// val 表示需要处理的正则表达式，比如 {id:\\d+} 则为 \\d+。
func Register(f MatchFunc, val ...string) error {
	if len(val) == 0 {
		panic("参数 val 不能为空")
	}

	for _, v := range val {
		if _, exists := interceptors.Load(v); exists {
			return fmt.Errorf("%s 已经存在", v)
		}

		interceptors.Store(v, f)
	}

	return nil
}

// Deregister 注销拦截器
func Deregister(val ...string) {
	for _, v := range val {
		interceptors.Delete(v)
	}
}

// Get 查找指定的处理函数
func Get(v string) (MatchFunc, bool) {
	f, found := interceptors.Load(v)
	if !found {
		return nil, false
	}
	return f.(MatchFunc), true
}
