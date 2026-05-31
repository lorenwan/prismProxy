package collection

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Runner 请求执行器
type Runner struct {
	client      *http.Client
	environment map[string]string
}

// NewRunner 创建新的请求执行器
func NewRunner() *Runner {
	return &Runner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		environment: make(map[string]string),
	}
}

// SetEnvironment 设置环境变量
func (r *Runner) SetEnvironment(env map[string]string) {
	r.environment = env
}

// Execute 执行请求
func (r *Runner) Execute(req *APIRequest) (*ExecutionResult, error) {
	if req == nil {
		return nil, fmt.Errorf("请求配置为空")
	}

	// 替换环境变量
	url := r.replaceVariables(req.URL)
	method := r.replaceVariables(req.Method)

	// 创建请求体
	var bodyReader io.Reader
	if req.Body != nil && req.Body.Content != "" {
		content := r.replaceVariables(req.Body.Content)
		bodyReader = strings.NewReader(content)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for _, h := range req.Headers {
		if h.Enabled {
			key := r.replaceVariables(h.Key)
			value := r.replaceVariables(h.Value)
			httpReq.Header.Set(key, value)
		}
	}

	// 设置查询参数
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for _, p := range req.QueryParams {
			if p.Enabled {
				key := r.replaceVariables(p.Key)
				value := r.replaceVariables(p.Value)
				q.Set(key, value)
			}
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// 设置认证
	if req.Auth != nil {
		r.applyAuth(httpReq, req.Auth)
	}

	// 设置默认 Content-Type
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		switch req.Body.Type {
		case "json":
			httpReq.Header.Set("Content-Type", "application/json")
		case "form":
			httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case "xml":
			httpReq.Header.Set("Content-Type", "application/xml")
		}
	}

	// 执行请求
	start := time.Now()
	resp, err := r.client.Do(httpReq)
	duration := time.Since(start)

	result := &ExecutionResult{
		RequestID: req.ID,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result, nil
	}

	// 填充结果
	result.Status = resp.StatusCode
	result.StatusText = resp.Status
	result.ContentType = resp.Header.Get("Content-Type")
	result.Size = int64(len(bodyBytes))
	result.Body = string(bodyBytes)

	// 收集响应头
	for key, values := range resp.Header {
		result.Headers = append(result.Headers, KeyValue{
			Key:   key,
			Value: strings.Join(values, ", "),
		})
	}

	// 执行测试
	if len(req.Tests) > 0 {
		result.TestResults = r.runTests(req.Tests, resp, result)
	}

	return result, nil
}

// replaceVariables 替换变量 {{variable}}
func (r *Runner) replaceVariables(input string) string {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		varName := strings.Trim(match, "{}")
		if value, ok := r.environment[varName]; ok {
			return value
		}
		return match
	})
}

// applyAuth 应用认证
func (r *Runner) applyAuth(req *http.Request, auth *AuthConfig) {
	switch auth.Type {
	case "basic":
		username := auth.Config["username"]
		password := auth.Config["password"]
		req.SetBasicAuth(username, password)
	case "bearer":
		token := auth.Config["token"]
		req.Header.Set("Authorization", "Bearer "+token)
	case "apikey":
		key := auth.Config["key"]
		value := auth.Config["value"]
		location := auth.Config["location"]
		if location == "header" {
			req.Header.Set(key, value)
		} else {
			q := req.URL.Query()
			q.Set(key, value)
			req.URL.RawQuery = q.Encode()
		}
	}
}

// runTests 执行测试
func (r *Runner) runTests(tests []Test, resp *http.Response, result *ExecutionResult) []TestResult {
	var results []TestResult

	for _, test := range tests {
		if !test.Enabled {
			continue
		}

		tr := TestResult{Test: test}

		switch test.Type {
		case "status":
			tr.Actual = fmt.Sprintf("%d", result.Status)
			tr.Passed = compareInt(result.Status, test.Operator, test.Value)
		case "header":
			actual := resp.Header.Get(test.Target)
			tr.Actual = actual
			tr.Passed = compareString(actual, test.Operator, test.Value)
		case "response_time":
			tr.Actual = fmt.Sprintf("%d", result.Duration.Milliseconds())
			tr.Passed = compareInt(int(result.Duration.Milliseconds()), test.Operator, test.Value)
		case "body":
			tr.Actual = result.Body
			tr.Passed = compareString(result.Body, test.Operator, test.Value)
		}

		results = append(results, tr)
	}

	return results
}

// compareInt 比较整数
func compareInt(actual int, operator, expected string) bool {
	var expectedInt int
	fmt.Sscanf(expected, "%d", &expectedInt)

	switch operator {
	case "eq":
		return actual == expectedInt
	case "ne":
		return actual != expectedInt
	case "gt":
		return actual > expectedInt
	case "lt":
		return actual < expectedInt
	case "gte":
		return actual >= expectedInt
	case "lte":
		return actual <= expectedInt
	}
	return false
}

// compareString 比较字符串
func compareString(actual, operator, expected string) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "ne":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case "not_contains":
		return !strings.Contains(actual, expected)
	case "starts_with":
		return strings.HasPrefix(actual, expected)
	case "ends_with":
		return strings.HasSuffix(actual, expected)
	}
	return false
}

// GenerateCode 生成代码（代理到 codegen 模块）
func (r *Runner) GenerateCode(req *APIRequest, language string) (string, error) {
	// 这里会调用 codegen 模块
	return "", fmt.Errorf("代码生成功能需要 codegen 模块")
}

// BuildCurlCommand 构建 cURL 命令
func BuildCurlCommand(req *APIRequest) string {
	var parts []string

	// 方法
	if req.Method != "" && req.Method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", req.Method))
	}

	// URL
	url := req.URL
	if len(req.QueryParams) > 0 {
		params := make([]string, 0, len(req.QueryParams))
		for _, p := range req.QueryParams {
			if p.Enabled {
				params = append(params, fmt.Sprintf("%s=%s", p.Key, p.Value))
			}
		}
		if len(params) > 0 {
			separator := "?"
			if strings.Contains(url, "?") {
				separator = "&"
			}
			url += separator + strings.Join(params, "&")
		}
	}

	// 请求头
	for _, h := range req.Headers {
		if h.Enabled {
			parts = append(parts, fmt.Sprintf("-H '%s: %s'", h.Key, h.Value))
		}
	}

	// 请求体
	if req.Body != nil && req.Body.Content != "" {
		parts = append(parts, fmt.Sprintf("-d '%s'", req.Body.Content))
	}

	// 认证
	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			parts = append(parts, fmt.Sprintf("-u '%s:%s'",
				req.Auth.Config["username"],
				req.Auth.Config["password"]))
		case "bearer":
			parts = append(parts, fmt.Sprintf("-H 'Authorization: Bearer %s'",
				req.Auth.Config["token"]))
		}
	}

	parts = append(parts, fmt.Sprintf("'%s'", url))

	return "curl " + strings.Join(parts, " ")
}

// NewHTTPRequest 创建 HTTP 请求（用于外部调用）
func NewHTTPRequest(req *APIRequest) (*http.Request, error) {
	var bodyReader io.Reader
	if req.Body != nil && req.Body.Content != "" {
		bodyReader = bytes.NewReader([]byte(req.Body.Content))
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	for _, h := range req.Headers {
		if h.Enabled {
			httpReq.Header.Set(h.Key, h.Value)
		}
	}

	return httpReq, nil
}
