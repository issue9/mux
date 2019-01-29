// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package host 管理路由中域名的切换等操作
package host

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/mux/v2/internal/handlers"
	"github.com/issue9/mux/v2/internal/tree"
	"github.com/issue9/mux/v2/params"
)

var errPatternNotFound = errors.New("pattern 不存在")

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

// New 声明新的 Hosts 变量
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
	domain, pattern := hs.split(pattern)
	return hs.getTree(domain).Add(pattern, h, method...)
}

// SetAllow 设置 Options 的 allow 报头值
func (hs *Hosts) SetAllow(pattern string, allow string) {
	domain, pattern := hs.split(pattern)
	hs.getTree(domain).SetAllow(pattern, allow)
}

// Remove 移除指定的路由项。
func (hs *Hosts) Remove(pattern string, method ...string) {
	domain, pattern := hs.split(pattern)

	if tree := hs.findTree(domain); tree != nil {
		tree.Remove(pattern, method...)
	}
}

// URL 根据参数生成地址。
func (hs *Hosts) URL(pattern string, params map[string]string) (string, error) {
	domain, pattern := hs.split(pattern)

	if tree := hs.findTree(domain); tree != nil {
		url, err := tree.URL(pattern, params)
		if err != nil {
			return "", err
		}
		return domain + url, nil
	}

	return "", errPatternNotFound
}

// CleanAll 清除所有的路由项
func (hs *Hosts) CleanAll() {
	for _, host := range hs.hosts {
		host.tree.Clean("")
	}
	hs.hosts = hs.hosts[:0]

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

	domain, pattern := hs.split(prefix)

	for _, host := range hs.hosts {
		if host.raw == domain {
			host.tree.Clean(pattern)
			return
		}
	}
}

// Handler 获取匹配的路由项
func (hs *Hosts) Handler(hostname, path string) (*handlers.Handlers, params.Params) {
	if !hs.skipCleanPath {
		path = cleanPath(path)
	}

	for _, host := range hs.hosts {
		if host.wildcard && strings.HasSuffix(hostname, host.domain) {
			return host.tree.Handler(path)
		}

		if hostname == host.domain {
			return host.tree.Handler(path)
		}
	}

	return hs.tree.Handler(path)
}

func (hs *Hosts) split(url string) (domain, pattern string) {
	if url == "" {
		panic("路由项地址不能为空")
	}

	if url[0] == '/' {
		return "", url
	}

	index := strings.IndexByte(url, '/')
	if index < 0 {
		panic(fmt.Errorf("%s 不能只指定域名部分", url))
	}

	return url[:index], url[index:]
}

// 获取指定路由项对应的 tree.Tree 实例，如果不存在，则返回空值。
func (hs *Hosts) findTree(domain string) *tree.Tree {
	if domain == "" {
		return hs.tree
	}

	for _, host := range hs.hosts {
		if host.raw == domain {
			return host.tree
		}
	}

	return nil
}

// 获取指定路由项对应的 tree.Tree 实例，如果不存在，则添加并返回。
func (hs *Hosts) getTree(domain string) *tree.Tree {
	if domain == "" {
		return hs.tree
	}

	for _, host := range hs.hosts {
		if host.raw == domain {
			return host.tree
		}
	}

	host := newHost(domain, tree.New(hs.disableOptions, hs.disableHead))

	// 对域名列表进行排序，非通配符版本在前面
	hs.hosts = append(hs.hosts, host)
	sort.SliceStable(hs.hosts, func(i, j int) bool {
		ii := hs.hosts[i]
		jj := hs.hosts[j]

		switch {
		case ii.wildcard == jj.wildcard: // 同为 true 或是 false
			return ii.domain < jj.domain
		case ii.wildcard:
			return false
		case jj.wildcard:
			return true
		default:
			return ii.domain < jj.domain
		}
	})

	return host.tree
}

func newHost(domain string, tree *tree.Tree) *host {
	host := &host{
		raw:    domain,
		domain: domain,
		tree:   tree,
	}

	if strings.HasPrefix(host.domain, "*.") {
		host.wildcard = true
		host.domain = domain[1:] // 保留 . 符号
	}

	return host
}
