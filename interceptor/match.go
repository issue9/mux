// SPDX-License-Identifier: MIT

package interceptor

// MatchFunc 每个拦截器的实际处理函数
type MatchFunc func(string) bool

func init() {
	if err := Register(MatchDigit, "digit"); err != nil {
		panic(err)
	}

	if err := Register(MatchDigit, "word"); err != nil {
		panic(err)
	}
}

// MatchAny 匹配任意字符
func MatchAny(path string) bool { return true }

// MatchDigit 匹配数值字符
//
// 正则表达式中的 [0-9]+ 是相同的。
func MatchDigit(path string) bool {
	for _, c := range path {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// MatchWord 匹配单词
//
// 正则表达式中的 [a-zA-Z0-9]+ 是相同的。
func MatchWord(path string) bool {
	for _, c := range path {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}
