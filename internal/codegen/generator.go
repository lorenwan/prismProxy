package codegen

import (
	"fmt"
	"strings"
)

// Generator 代码生成器
type Generator struct {
	templates map[Language]*Template
}

// NewGenerator 创建新的代码生成器
func NewGenerator() *Generator {
	return &Generator{
		templates: templates,
	}
}

// Generate 生成代码
func (g *Generator) Generate(language Language, req *RequestData) (string, error) {
	tmpl, ok := g.templates[language]
	if !ok {
		return "", fmt.Errorf("不支持的语言: %s", language)
	}

	if req == nil {
		return "", fmt.Errorf("请求数据为空")
	}

	return tmpl.Generate(req), nil
}

// GetSupportedLanguages 获取支持的语言列表
func (g *Generator) GetSupportedLanguages() []map[string]string {
	var languages []map[string]string
	for _, tmpl := range g.templates {
		languages = append(languages, map[string]string{
			"id":          string(tmpl.Language),
			"name":        tmpl.Name,
			"description": tmpl.Description,
		})
	}
	return languages
}

// FromCollectionRequest 从集合请求转换为请求数据
func FromCollectionRequest(method, url string, headers []KeyValue, body *RequestBody, auth *AuthData) *RequestData {
	req := &RequestData{
		Method: method,
		URL:    url,
	}

	// 转换请求头
	req.Headers = make(map[string]string)
	for _, h := range headers {
		req.Headers[h.Key] = h.Value
	}

	// 转换请求体
	if body != nil {
		req.Body = body.Content
		req.BodyType = body.Type
	}

	// 转换认证
	req.Auth = auth

	return req
}

// KeyValue 键值对（codegen 内部使用）
type KeyValue struct {
	Key   string
	Value string
}

// RequestBody 请求体（codegen 内部使用）
type RequestBody struct {
	Type    string
	Content string
}

// ParseURL 解析 URL，提取查询参数
func ParseURL(url string) (string, map[string]string) {
	params := make(map[string]string)

	parts := strings.SplitN(url, "?", 2)
	if len(parts) < 2 {
		return url, params
	}

	baseURL := parts[0]
	queryString := parts[1]

	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	return baseURL, params
}

// BuildURL 构建 URL，添加查询参数
func BuildURL(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	var pairs []string
	for key, value := range params {
		pairs = append(pairs, key+"="+value)
	}

	return baseURL + "?" + strings.Join(pairs, "&")
}

// FormatJSON 格式化 JSON（简单实现）
func FormatJSON(json string) string {
	// 简单的 JSON 格式化
	var result strings.Builder
	indent := 0
	inString := false
	escaped := false

	for _, ch := range json {
		if escaped {
			result.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			result.WriteRune(ch)
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			result.WriteRune(ch)
			continue
		}

		if inString {
			result.WriteRune(ch)
			continue
		}

		switch ch {
		case '{', '[':
			result.WriteRune(ch)
			indent++
			result.WriteString("\n" + strings.Repeat("  ", indent))
		case '}', ']':
			indent--
			result.WriteString("\n" + strings.Repeat("  ", indent))
			result.WriteRune(ch)
		case ',':
			result.WriteRune(ch)
			result.WriteString("\n" + strings.Repeat("  ", indent))
		case ':':
			result.WriteString(": ")
		default:
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// GenerateFromTraffic 从流量数据生成代码
func (g *Generator) GenerateFromTraffic(language Language, trafficID int64) (string, error) {
	// 这里需要从存储层获取流量数据
	// 暂时返回错误，需要集成存储层
	return "", fmt.Errorf("需要集成存储层")
}

// GenerateBatch 批量生成代码
func (g *Generator) GenerateBatch(languages []Language, req *RequestData) (map[Language]string, error) {
	results := make(map[Language]string)

	for _, lang := range languages {
		code, err := g.Generate(lang, req)
		if err != nil {
			return nil, fmt.Errorf("生成 %s 代码失败: %w", lang, err)
		}
		results[lang] = code
	}

	return results, nil
}
