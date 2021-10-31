// SPDX-License-Identifier: MIT

package group

import (
	"net/http"

	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

// Hosts 限定域名的匹配工具
type Hosts struct {
	tree *tree.Tree
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(domain ...string) *Hosts {
	h := &Hosts{tree: tree.New(false, false)}
	h.Add(domain...)
	return h
}

func (hs *Hosts) Match(r *http.Request) (*http.Request, bool) {
	hostname := r.URL.Hostname()
	h, ps := hs.tree.Route(hostname)
	if h == nil {
		return nil, false
	}

	return params.WithValue(r, ps), true
}

// Add 添加新的域名
//
// 域名的格式和路由的语法格式是一样的，比如：
//  api.example.com
//  {sub:[a-z]+}.example.com
// 如果存在命名参数，也可以通过 params.Get 获取。
// 当语法错误时，会触发 panic，可通过 CheckSyntax 检测语法的正确性。
func (hs *Hosts) Add(domain ...string) {
	for _, d := range domain {
		err := hs.tree.Add(d, http.HandlerFunc(hs.emptyHandlerFunc), http.MethodGet)
		if err != nil {
			panic(err)
		}
	}
}

// Delete 删除域名
func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc(http.ResponseWriter, *http.Request) {}
