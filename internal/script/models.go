package script

import "time"

// Script 脚本定义
type Script struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Content   string       `json:"content"`
	Phase     ScriptPhase  `json:"phase"`
	Enabled   bool         `json:"enabled"`
	Priority  int          `json:"priority"`
	Language  ScriptType   `json:"language"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// ScriptType 脚本类型
type ScriptType string

const (
	ScriptTypeExpr ScriptType = "expr"
)

// ScriptPhase 脚本执行阶段
type ScriptPhase string

const (
	PhaseRequest  ScriptPhase = "request"
	PhaseResponse ScriptPhase = "response"
)

// ScriptExecution 脚本执行记录
type ScriptExecution struct {
	ScriptID      string    `json:"script_id"`
	TransactionID string    `json:"transaction_id"`
	Success       bool      `json:"success"`
	Output        string    `json:"output"`
	Duration      int64     `json:"duration_ms"`
	Error         string    `json:"error,omitempty"`
	ExecutedAt    time.Time `json:"executed_at"`
}
