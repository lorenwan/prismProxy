package debugger

import (
	"time"

	"prismproxy/internal/rules"
	"prismproxy/internal/traffic"
)

// Phase 断点阶段
type Phase string

const (
	PhaseRequest  Phase = "request"
	PhaseResponse Phase = "response"
)

// BreakActionType 断点动作类型
type BreakActionType string

const (
	BreakActionPause       BreakActionType = "pause"
	BreakActionAutoModify  BreakActionType = "auto_modify"
	BreakActionDrop        BreakActionType = "drop"
)

// Breakpoint 断点定义
type Breakpoint struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Enabled   bool                `json:"enabled"`
	Phase     Phase               `json:"phase"`
	Match     rules.RuleMatch     `json:"match"`
	Action    BreakAction         `json:"action"`
	HitCount  int                 `json:"hit_count"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// BreakAction 断点动作
type BreakAction struct {
	Type          BreakActionType   `json:"type"`
	Modifications *rules.ModifySpec `json:"modifications,omitempty"`
}

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionStatusPaused   SessionStatus = "paused"
	SessionStatusModified SessionStatus = "modified"
	SessionStatusReleased SessionStatus = "released"
	SessionStatusDropped  SessionStatus = "dropped"
)

// BreakpointSession 断点会话
type BreakpointSession struct {
	ID            string               `json:"id"`
	BreakpointID  string               `json:"breakpoint_id"`
	TransactionID int64                `json:"transaction_id"`
	Phase         Phase                `json:"phase"`
	Status        SessionStatus        `json:"status"`
	Original      *traffic.Transaction `json:"original"`
	Modified      *traffic.Transaction `json:"modified,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	ResolvedAt    *time.Time           `json:"resolved_at,omitempty"`
}
