package environment

import "time"

// Environment 环境
type Environment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Variables []Variable `json:"variables,omitempty"`
	IsActive  bool       `json:"is_active"`
	IsDefault bool       `json:"is_default"`
	BaseURL   string     `json:"base_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Variable 环境变量
type Variable struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	IsSecret    bool   `json:"is_secret"`
}

// EnvironmentExport 导出格式
type EnvironmentExport struct {
	Version     string            `json:"version"`
	Name        string            `json:"name"`
	Variables   []VariableExport  `json:"variables"`
}

// VariableExport 变量导出格式
type VariableExport struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
}
