package rules

import (
	"net/http"
	"regexp"
	"strings"
)

// Matcher 规则匹配器
type Matcher struct{}

// NewMatcher 创建新的匹配器
func NewMatcher() *Matcher {
	return &Matcher{}
}

// MatchRequest 匹配请求
func (m *Matcher) MatchRequest(rule *Rule, method, url, host string, headers http.Header) bool {
	if !rule.Enabled {
		return false
	}

	match := rule.Match

	// 匹配方法
	if len(match.Methods) > 0 {
		methodMatched := false
		for _, m := range match.Methods {
			if strings.EqualFold(m, method) {
				methodMatched = true
				break
			}
		}
		if !methodMatched {
			return false
		}
	}

	// 匹配主机
	if match.HostPattern != "" {
		if !m.matchPattern(match.HostPattern, host) {
			return false
		}
	}

	// 匹配 URL
	if match.URLPattern != "" {
		matched, err := regexp.MatchString(match.URLPattern, url)
		if err != nil || !matched {
			return false
		}
	}

	if match.URLWildcard != "" {
		if !m.matchWildcard(match.URLWildcard, url) {
			return false
		}
	}

	// 匹配 Header
	if match.HeaderMatch != nil {
		if !m.matchHeader(match.HeaderMatch, headers) {
			return false
		}
	}

	// 匹配 Content-Type
	if len(match.ContentType) > 0 {
		ct := headers.Get("Content-Type")
		ctMatched := false
		for _, ctPattern := range match.ContentType {
			if strings.Contains(ct, ctPattern) {
				ctMatched = true
				break
			}
		}
		if !ctMatched {
			return false
		}
	}

	return true
}

// matchPattern 匹配模式（支持正则）
func (m *Matcher) matchPattern(pattern, value string) bool {
	// 尝试正则匹配
	matched, err := regexp.MatchString(pattern, value)
	if err == nil {
		return matched
	}

	// 回退到字符串包含
	return strings.Contains(value, pattern)
}

// matchWildcard 匹配通配符
func (m *Matcher) matchWildcard(pattern, value string) bool {
	// 将通配符转换为正则
	regexPattern := "^" + strings.ReplaceAll(pattern, "*", ".*") + "$"
	matched, err := regexp.MatchString(regexPattern, value)
	if err != nil {
		return false
	}
	return matched
}

// matchHeader 匹配 Header
func (m *Matcher) matchHeader(hm *HeaderMatch, headers http.Header) bool {
	values := headers.Values(hm.Name)
	if len(values) == 0 {
		return false
	}

	for _, v := range values {
		switch hm.MatchType {
		case "exact":
			if v == hm.Value {
				return true
			}
		case "contains":
			if strings.Contains(v, hm.Value) {
				return true
			}
		case "regex":
			matched, err := regexp.MatchString(hm.Value, v)
			if err == nil && matched {
				return true
			}
		default:
			if strings.Contains(v, hm.Value) {
				return true
			}
		}
	}

	return false
}

// MatchRules 匹配所有规则，返回匹配的规则列表
func (m *Matcher) MatchRules(rules []*Rule, method, url, host string, headers http.Header) []*Rule {
	var matched []*Rule

	for _, rule := range rules {
		if m.MatchRequest(rule, method, url, host, headers) {
			matched = append(matched, rule)
		}
	}

	return matched
}

// SortByPriority 按优先级排序
func SortByPriority(rules []*Rule) []*Rule {
	// 简单的冒泡排序，优先级数字越小越优先
	sorted := make([]*Rule, len(rules))
	copy(sorted, rules)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority > sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}
