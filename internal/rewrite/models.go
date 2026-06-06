package rewrite

import (
	"time"

	"prismproxy/internal/rules"
)

// RewriteType 重写类型
type RewriteType string

const (
	RewriteAddHeader     RewriteType = "add_header"
	RewriteRemoveHeader  RewriteType = "remove_header"
	RewriteReplaceHeader RewriteType = "replace_header"
	RewriteReplaceBody   RewriteType = "replace_body"
	RewriteReplaceURL    RewriteType = "replace_url"
	RewriteMapLocal      RewriteType = "map_local"
	RewriteMapRemote     RewriteType = "map_remote"
)

// RewriteWhere 重写位置
type RewriteWhere string

const (
	RewriteWhereRequestHeader  RewriteWhere = "request_header"
	RewriteWhereResponseHeader RewriteWhere = "response_header"
	RewriteWhereRequestBody    RewriteWhere = "request_body"
	RewriteWhereResponseBody   RewriteWhere = "response_body"
	RewriteWhereURLQuery       RewriteWhere = "url_query"
	RewriteWhereURLPath        RewriteWhere = "url_path"

	// 兼容别名
	RewriteWhereRequest  = RewriteWhereRequestHeader
	RewriteWhereResponse = RewriteWhereResponseHeader
)

// IsRequest 判断是否为请求阶段的重写位置
func (w RewriteWhere) IsRequest() bool {
	switch w {
	case RewriteWhereRequestHeader, RewriteWhereRequestBody, RewriteWhereURLQuery, RewriteWhereURLPath:
		return true
	default:
		return false
	}
}

// IsResponse 判断是否为响应阶段的重写位置
func (w RewriteWhere) IsResponse() bool {
	switch w {
	case RewriteWhereResponseHeader, RewriteWhereResponseBody:
		return true
	default:
		return false
	}
}

// RewriteRule 重写规则
type RewriteRule struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	Priority  int             `json:"priority"`
	Match     rules.RuleMatch `json:"match"`
	Actions   []RewriteAction `json:"actions"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// RewriteAction 重写动作
type RewriteAction struct {
	Type   RewriteType  `json:"type"`
	Where  RewriteWhere `json:"where"`
	Key    string       `json:"key,omitempty"`    // header name 等
	Value  string       `json:"value,omitempty"`  // 替换值
	Target string       `json:"target,omitempty"` // 本地文件路径或远程 URL
}
