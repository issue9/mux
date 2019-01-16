// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package host 管理路由中域名的切换等操作
package host

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/mux/v2/internal/handlers"
	"github.com/issue9/mux/v2/internal/tree"
	"github.com/issue9/mux/v2/params"
)

type host struct {
	raw      string // 域名的原始值，非通配符版本，与 domain 相同
	domain   string // 域名
	wildcard bool   // 是否带通配符

	tree *tree.Tree
}

// Hosts 域名管理
type Hosts struct {
	disableOptions bool
	disableHead    bool
	skipCleanPath  bool
	hosts          []*host
	tree           *tree.Tree // 非域名限定的路由项
}

// New 声明新的 Host 变量
func New(disableOptions, disableHead, skipCleanPath bool) *Hosts {
	return &Hosts{
		disableHead:    disableHead,
		disableOptions: disableOptions,
		skipCleanPath:  skipCleanPath,
		hosts:          make([]*host, 0, 10),
		tree:           tree.New(disableOptions, disableHead),
	}
}

// Add 添加路由项
func (hs *Hosts) Add(pattern string, h http.Handler, method ...string) error {
	tree, err := hs.getTree(pattern)
	if err != nil {
		return err
	}

	return tree.Add(pattern, h, method...)
}

// SetAllow 设置 Options 的 allow 报头值
func (hs *Hosts) SetAllow(pattern string, allow string) error {
	tree, err := hs.getTree(pattern)
	if err != nil {
		return err
	}

	return tree.SetAllow(pattern, allow)
}

// Remove 移除指定的路由项。
func (hs *Hosts) Remove(pattern string, method ...string) {
	tree, err := hs.findTree(pattern)
	if err != nil {
		panic(err)
	}

	tree.Remove(pattern, method...)
}

// URL 根据参数生成地址。
func (hs *Hosts) URL(pattern string, params map[string]string) (string, error) {
	tree, err := hs.findTree(pattern)
	if err != nil {
		return "", err
	}

	return tree.URL(pattern, params)
}

// CleanAll 清除所有的路由项
func (hs *Hosts) CleanAll() {
	for _, host := range hs.hosts {
		host.tree.Clean("")
	}
	hs.tree.Clean("")
}

// Clean 消除指定前缀的路由项
func (hs *Hosts) Clean(prefix string) {
	if prefix == "" {
		hs.CleanAll()
		return
	}

	if prefix[0] == '/' {
		hs.tree.Clean(prefix)
		return
	}

	index := strings.IndexByte(prefix, '/')
	if index < 0 {
		panic(fmt.Errorf("%s 不能只指定域名部分", prefix))
	}

	domain := prefix[:index]

	for _, host := range hs.hosts {
		if host.raw == domain {
			host.tree.Clean(prefix[index:])
			return
		}
	}
}

// Handler 获取匹配的路由项
func (hs *Hosts) Handler(r *http.Request) (*handlers.Handlers, params.Params) {
	p := r.URL.Path
	if !hs.skipCleanPath {
		p = cleanPath(p)
	}

	for _, host := range hs.hosts {
		if host.wildcard && strings.HasSuffix(r.Host, host.domain) {
			return host.tree.Handler(p)
		}

		if r.Host == host.domain {
			return host.tree.Handler(p)
		}
	}

	return hs.tree.Handler(p)
}

// 获取指定路由项对应的 tree.Tree 实例，如果不存在，则返回空值。
func (hs *Hosts) findTree(pattern string) (*tree.Tree, error) {
	if pattern == "" {
		panic("路由项地址不能为空")
	}

	if pattern[0] == '/' {
		return hs.tree, nil
	}

	index := strings.IndexByte(pattern, '/')
	if index < 0 {
		return nil, fmt.Errorf("%s 不能只指定域名部分", pattern)
	}

	domain := pattern[:index]

	for _, host := range hs.hosts {
		if host.raw == domain {
			return host.tree, nil
		}
	}

	return nil, nil
}

// 获取指定路由项对应的 tree.Tree 实例，如果不存在，则添加并返回。
func (hs *Hosts) getTree(pattern string) (*tree.Tree, error) {
	if pattern == "" {
		panic("路由项地址不能为空")
	}

	if pattern[0] == '/' {
		return hs.tree, nil
	}

	index := strings.IndexByte(pattern, '/')
	if index < 0 {
		return nil, fmt.Errorf("%s 不能只指定域名部分", pattern)
	}

	domain := pattern[:index]

	for _, host := range hs.hosts {
		if host.raw == domain {
			return host.tree, nil
		}
	}

	host := &host{
		raw:    domain,
		domain: domain,
		tree:   tree.New(hs.disableOptions, hs.disableHead),
	}

	if strings.HasPrefix(host.domain, "*.") {
		host.wildcard = true
		host.domain = domain[1:] // 保留 . 符号
	}

	// 对域名列表进行排序，非通配符版本在前面
	hs.hosts = append(hs.hosts, host)
	sort.SliceStable(hs.hosts, func(i, j int) bool {
		ii := hs.hosts[i]
		jj := hs.hosts[j]

		switch {
		case ii.wildcard:
			return true
		case jj.wildcard:
			return true
		default:
			return ii.domain < jj.domain
		}
	})

	return host.tree, nil
}

// 清除路径中的重复的 / 字符
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	index := strings.Index(p, "//")
	if index == -1 {
		return p
	}

	pp := make([]byte, index+1, len(p))
	copy(pp, p[:index+1])

	slash := true
	for i := index + 2; i < len(p); i++ {
		if p[i] == '/' {
			if slash {
				continue
			}
			slash = true
		} else {
			slash = false
		}
		pp = append(pp, p[i])
	}

	return string(pp)
}
