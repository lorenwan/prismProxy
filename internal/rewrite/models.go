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
	RewriteWhereRequest  RewriteWhere = "request"
	RewriteWhereResponse RewriteWhere = "response"
)

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
	Type   RewriteType `json:"type"`
	Where  RewriteWhere `json:"where"`
	Key    string      `json:"key,omitempty"`    // header name 等
	Value  string      `json:"value,omitempty"`  // 替换值
	Target string      `json:"target,omitempty"` // 本地文件路径或远程 URL
}
