// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
	"github.com/issue9/mux/v6/params"
)

// Hosts 限定域名的匹配工具
type Hosts struct {
	i    *syntax.Interceptors
	tree *tree.Tree
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(lock bool, domain ...string) *Hosts {
	i := syntax.NewInterceptors()
	h := &Hosts{tree: tree.New(lock, i), i: i}
	h.Add(domain...)
	return h
}

// RegisterInterceptor 注册拦截器
//
// NOTE: 拦截器只有在注册之后添加的域名才有效果。
func (hs *Hosts) RegisterInterceptor(f mux.InterceptorFunc, name ...string) {
	hs.i.Add(f, name...)
}

func (hs *Hosts) Match(r *http.Request) (params.Params, bool) {
	h, ps := hs.tree.Route(strings.ToLower(r.URL.Hostname()))
	if h == nil || h.Handler(http.MethodGet) == nil {
		return nil, false
	}
	return ps, true
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
		err := hs.tree.Add(d, hs.emptyHandlerFunc, http.MethodGet)
		if err != nil {
			panic(err)
		}
	}
}

// Delete 删除域名
func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc(http.ResponseWriter, *http.Request, params.Params) {}
