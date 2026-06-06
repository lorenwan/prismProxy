package rules

import (
	"time"
)

// Rule 规则定义
type Rule struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Enabled   bool       `json:"enabled"`
	Priority  int        `json:"priority"`
	Match     RuleMatch  `json:"match"`
	Action    RuleAction `json:"action"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// RuleMatch 匹配条件
type RuleMatch struct {
	// URL 匹配
	URLPattern  string `json:"url_pattern,omitempty"`
	URLWildcard string `json:"url_wildcard,omitempty"`

	// 主机匹配
	HostPattern string `json:"host_pattern,omitempty"`

	// 方法匹配
	Methods []string `json:"methods,omitempty"`

	// Header 匹配
	HeaderMatch *HeaderMatch `json:"header_match,omitempty"`

	// Content-Type 匹配
	ContentType []string `json:"content_type,omitempty"`
}

// HeaderMatch Header 匹配规则
type HeaderMatch struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	MatchType string `json:"match_type"` // exact, contains, regex
}

// RuleAction 规则动作
type RuleAction struct {
	Type ActionType `json:"type"`

	// Map Local 相关
	LocalPath string `json:"local_path,omitempty"`

	// Map Remote 相关
	RemoteURL string `json:"remote_url,omitempty"`

	// 修改请求/响应
	Modify *ModifySpec `json:"modify,omitempty"`

	// 阻止请求
	BlockResponse *BlockSpec `json:"block_response,omitempty"`

	// 延迟
	DelayMs int `json:"delay_ms,omitempty"`
}

// ActionType 动作类型
type ActionType string

const (
	ActionMapLocal       ActionType = "map_local"
	ActionMapRemote      ActionType = "map_remote"
	ActionModifyRequest  ActionType = "modify_request"
	ActionModifyResponse ActionType = "modify_response"
	ActionBlock          ActionType = "block"
	ActionDelay          ActionType = "delay"
	ActionMock           ActionType = "mock"
)

// ModifySpec 修改规格
type ModifySpec struct {
	AddHeaders    map[string]string `json:"add_headers,omitempty"`
	RemoveHeaders []string          `json:"remove_headers,omitempty"`
	SetHeaders    map[string]string `json:"set_headers,omitempty"`
	AddQuery      map[string]string `json:"add_query,omitempty"`
	RemoveQuery   []string          `json:"remove_query,omitempty"`
	SetQuery      map[string]string `json:"set_query,omitempty"`
	BodyReplace   string            `json:"body_replace,omitempty"`
}

// BlockSpec 阻止规格
type BlockSpec struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

// RuleStats 规则统计
type RuleStats struct {
	TotalRules    int            `json:"total_rules"`
	EnabledRules  int            `json:"enabled_rules"`
	DisabledRules int            `json:"disabled_rules"`
	HitCounts     map[string]int `json:"hit_counts"`
}
