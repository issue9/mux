// SPDX-License-Identifier: MIT

package group

import (
	"context"
	"net/http"

	"github.com/issue9/mux/v4/internal/tree"
	"github.com/issue9/mux/v4/params"
)

// Hosts 限定域名的匹配工具
type Hosts struct {
	tree *tree.Tree
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(domain ...string) (*Hosts, error) {
	h := &Hosts{
		tree: tree.New(false, false),
	}

	if err := h.Add(domain...); err != nil {
		return nil, err
	}
	return h, nil
}

func (hs *Hosts) Match(r *http.Request) (*http.Request, bool) {
	hostname := r.URL.Hostname()
	h, ps := hs.tree.Handler(hostname)
	if h == nil {
		return nil, false
	}

	if len(ps) > 0 {
		r = r.WithContext(context.WithValue(r.Context(), params.ContextKeyParams, ps))
	}
	return r, true
}

// Add 添加新的域名
//
// 域名的格式和路由的语法格式是一样的，比如：
//  api.example.com
//  {sub:[a-z]+}.example.com
func (hs *Hosts) Add(domain ...string) error {
	for _, d := range domain {
		err := hs.tree.Add(d, http.HandlerFunc(hs.emptyHandlerFunc), http.MethodGet)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除域名
func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc(http.ResponseWriter, *http.Request) {}
