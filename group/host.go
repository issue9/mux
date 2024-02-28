// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/internal/tree"
	"github.com/issue9/mux/v7/types"
)

// Hosts 限定域名的匹配工具
type Hosts struct {
	i    *syntax.Interceptors
	tree *tree.Tree[any]
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(lock bool, domain ...string) *Hosts {
	i := syntax.NewInterceptors()
	f := func(types.Node) any { return nil }
	t := tree.New(lock, i, nil, f, f)
	h := &Hosts{tree: t, i: i}
	h.Add(domain...)
	return h
}

// RegisterInterceptor 注册拦截器
//
// NOTE: 拦截器只有在注册之后添加的域名才有效果。
func (hs *Hosts) RegisterInterceptor(f mux.InterceptorFunc, name ...string) {
	hs.i.Add(f, name...)
}

func (hs *Hosts) Match(r *http.Request, ctx *types.Context) bool {
	h := r.Host // r.URL.Hostname() 可能为空，r.Host 一直有值！
	if i := strings.LastIndexByte(h, ':'); i != -1 && validOptionalPort(h[i:]) {
		h = h[:i]
	}
	if strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]") { // ipv6
		h = h[1 : len(h)-1]
	}

	ctx.Path = strings.ToLower(h)
	_, _, exists := hs.tree.Handler(ctx, http.MethodGet)
	return exists
}

// 源自 https://github.com/golang/go/blob/d8762b2f4532cc2e5ec539670b88bbc469a13938/src/net/url/url.go#L769
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// Add 添加新的域名
//
// 域名的格式和路由的语法格式是一样的，比如：
//
//	api.example.com
//	{sub:[a-z]+}.example.com
//
// 如果存在命名参数，也可以通过也可通过 [types.Params] 接口获取。
// 当语法错误时，会触发 panic，可通过 [mux.CheckSyntax] 检测语法的正确性。
func (hs *Hosts) Add(domain ...string) {
	for _, d := range domain {
		err := hs.tree.Add(strings.ToLower(d), hs.emptyHandlerFunc, http.MethodGet)
		if err != nil {
			panic(err)
		}
	}
}

func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc() {}
