// SPDX-License-Identifier: MIT

package muxutil

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
	"github.com/issue9/mux/v6/types"
)

// Hosts 限定域名的匹配工具
type Hosts struct {
	i    *syntax.Interceptors
	tree *tree.Tree[any]
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(lock bool, domain ...string) *Hosts {
	i := syntax.NewInterceptors()
	h := &Hosts{tree: tree.New(lock, i, func(o types.Node) any { return nil }), i: i}
	h.Add(domain...)
	return h
}

// RegisterInterceptor 注册拦截器
//
// NOTE: 拦截器只有在注册之后添加的域名才有效果。
func (hs *Hosts) RegisterInterceptor(f mux.InterceptorFunc, name ...string) {
	hs.i.Add(f, name...)
}

func (hs *Hosts) Match(r *http.Request) (types.Params, bool) {
	// r.URL.Hostname() 可能为空，r.Host 一直有值！
	host := r.Host
	if index := strings.LastIndexByte(host, ':'); index != -1 && validOptionalPort(host[index:]) {
		host = host[:index]
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") { // ipv6
		host = host[1 : len(host)-1]
	}

	h, ps := hs.tree.Match(strings.ToLower(host))
	if h == nil {
		return nil, false
	}
	if _, exists := h.Handler(http.MethodGet); !exists {
		return nil, false
	}
	return ps, true
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
//  api.example.com
//  {sub:[a-z]+}.example.com
// 如果存在命名参数，也可以通过 syntax.GetParams 获取。
// 当语法错误时，会触发 panic，可通过 CheckSyntax 检测语法的正确性。
func (hs *Hosts) Add(domain ...string) {
	for _, d := range domain {
		err := hs.tree.Add(strings.ToLower(d), hs.emptyHandlerFunc, http.MethodGet)
		if err != nil {
			panic(err)
		}
	}
}

// Delete 删除域名
func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc() {}
