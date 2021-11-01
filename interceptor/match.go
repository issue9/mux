// SPDX-License-Identifier: MIT

package interceptor

// MatchFunc 每个拦截器的实际处理函数
type MatchFunc func(string) bool

func init() {
	Register(MatchDigit, "digit")
	Register(MatchWord, "word")
	Register(MatchAny, "any")
}

// MatchAny 匹配任意非空内容
func MatchAny(path string) bool { return len(path) > 0 }

// MatchDigit 匹配数值字符
//
// 与正则表达式中的 [0-9]+ 是相同的。
func MatchDigit(path string) bool {
	for _, c := range path {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(path) > 0
}

// MatchWord 匹配单词
//
// 与正则表达式中的 [a-zA-Z0-9]+ 是相同的。
func MatchWord(path string) bool {
	for _, c := range path {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return len(path) > 0
}
