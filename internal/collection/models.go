package collection

import "time"

// Collection API 集合
type Collection struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ParentID    string           `json:"parent_id,omitempty"`
	Items       []CollectionItem `json:"items,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// CollectionItem 集合项目（文件夹或请求）
type CollectionItem struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"` // folder, request
	Name      string           `json:"name"`
	Request   *APIRequest      `json:"request,omitempty"`
	Items     []CollectionItem `json:"items,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// APIRequest API 请求定义
type APIRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     []KeyValue        `json:"headers,omitempty"`
	QueryParams []KeyValue        `json:"query_params,omitempty"`
	Body        *RequestBody      `json:"body,omitempty"`
	Auth        *AuthConfig       `json:"auth,omitempty"`
	Tests       []Test            `json:"tests,omitempty"`
	Variables   []KeyValue        `json:"variables,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// KeyValue 键值对
type KeyValue struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// RequestBody 请求体
type RequestBody struct {
	Type    string       `json:"type"` // none, form, json, xml, raw, binary, graphql
	Content string       `json:"content,omitempty"`
	Binary  string       `json:"binary,omitempty"` // 文件路径
	GraphQL *GraphQLBody `json:"graphql,omitempty"`
}

// GraphQLBody GraphQL 请求体
type GraphQLBody struct {
	Query     string `json:"query"`
	Variables string `json:"variables,omitempty"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type   string            `json:"type"` // none, basic, bearer, apikey, oauth2
	Config map[string]string `json:"config,omitempty"`
}

// Test 测试脚本
type Test struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // status, header, body, jsonpath, response_time
	Target   string `json:"target,omitempty"`
	Operator string `json:"operator"` // eq, ne, gt, lt, contains, matches
	Value    string `json:"value"`
	Enabled  bool   `json:"enabled"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	RequestID   string        `json:"request_id"`
	Status      int           `json:"status"`
	StatusText  string        `json:"status_text"`
	Headers     []KeyValue    `json:"headers"`
	Body        string        `json:"body"`
	ContentType string        `json:"content_type"`
	Duration    time.Duration `json:"duration"`
	Size        int64         `json:"size"`
	Error       string        `json:"error,omitempty"`
	TestResults []TestResult  `json:"test_results,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// TestResult 测试结果
type TestResult struct {
	Test   Test   `json:"test"`
	Passed bool   `json:"passed"`
	Actual string `json:"actual,omitempty"`
	Error  string `json:"error,omitempty"`
}
