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

	"github.com/issue9/mux/v2/internal/tree"
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
	hosts          []*host
	tree           *tree.Tree // 非域名限定的路由项
}

// New 声明新的 Host 变量
func New(disableOptions, disableHead bool) *Hosts {
	return &Hosts{
		disableHead:    disableHead,
		disableOptions: disableOptions,
		hosts:          make([]*host, 0, 10),
		tree:           tree.New(disableOptions, disableHead),
	}
}

// Add 添加路由项
func (hs *Hosts) Add(path string, h http.Handler, method ...string) error {
	tree, err := hs.getTree(path)
	if err != nil {
		return err
	}

	return tree.Add(path, h, method...)
}

// SetAllow 设置 Options 的 allow 报头值
func (hs *Hosts) SetAllow(path string, allow string) error {
	tree, err := hs.getTree(path)
	if err != nil {
		return err
	}

	return tree.SetAllow(path, allow)
}

// Remove 移除指定的路由项。
func (hs *Hosts) Remove(path string, method ...string) {
	tree, err := hs.getTree(path)
	if err != nil {
		panic(err)
	}

	tree.Remove(path, method...)
}

// URL 根据参数生成地址。
func (hs *Hosts) URL(path string, params map[string]string) (string, error) {
	tree, err := hs.getTree(path)
	if err != nil {
		return "", err
	}

	return tree.URL(path, params)
}

// CleanAll 清除所有的路由项
func (hs *Hosts) CleanAll() {
	for _, host := range hs.hosts {
		host.tree.Clean("")
	}
	hs.tree.Clean("")
}

func (hs *Hosts) getTree(path string) (*tree.Tree, error) {
	if path == "" {
		panic("路由项地址不能为空")
	}

	if path[0] == '/' {
		return hs.tree, nil
	}

	index := strings.IndexByte(path, '/')
	if index < 0 {
		return nil, fmt.Errorf("%s 不能只指定域名部分", path)
	}

	domain := path[:index]

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
