package script

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ScriptEngine 脚本引擎
type ScriptEngine struct {
	functions map[string]interface{}
}

// NewEngine 创建新的脚本引擎
func NewEngine() *ScriptEngine {
	e := &ScriptEngine{
		functions: make(map[string]interface{}),
	}
	e.registerBuiltinFunctions()
	return e
}

// registerBuiltinFunctions 注册内置函数
func (e *ScriptEngine) registerBuiltinFunctions() {
	e.functions["json_path"] = jsonPath
	e.functions["base64_encode"] = base64Encode
	e.functions["base64_decode"] = base64Decode
	e.functions["md5"] = md5Hash
	e.functions["sha256"] = sha256Hash
	e.functions["timestamp"] = timestamp
	e.functions["uuid"] = generateUUID
	e.functions["regex_match"] = regexMatch
	e.functions["regex_replace"] = regexReplace
	e.functions["url_encode"] = urlEncode
	e.functions["url_decode"] = urlDecode
}

// Execute 执行脚本
func (e *ScriptEngine) Execute(ctx context.Context, script *Script, phase ScriptPhase, data map[string]interface{}) (*ScriptExecution, error) {
	if !script.Enabled {
		return nil, fmt.Errorf("脚本已禁用")
	}

	if script.Phase != phase {
		return nil, fmt.Errorf("脚本阶段不匹配: 期望 %s, 实际 %s", script.Phase, phase)
	}

	start := time.Now()
	exec := &ScriptExecution{
		ScriptID:   script.ID,
		ExecutedAt: start,
	}

	// 执行表达式
	output, err := e.evaluate(script.Content, data)
	exec.Duration = time.Since(start).Milliseconds()

	if err != nil {
		exec.Success = false
		exec.Error = err.Error()
		return exec, err
	}

	exec.Success = true
	exec.Output = fmt.Sprintf("%v", output)
	return exec, nil
}

// evaluate 简单表达式求值
func (e *ScriptEngine) evaluate(expr string, data map[string]interface{}) (interface{}, error) {
	// 简单的表达式解析和执行
	expr = strings.TrimSpace(expr)

	// 处理函数调用
	if idx := strings.Index(expr, "("); idx > 0 {
		funcName := strings.TrimSpace(expr[:idx])
		if fn, ok := e.functions[funcName]; ok {
			// 提取参数
			args := extractArgs(expr[idx+1:])
			return callFunction(fn, args, data)
		}
	}

	// 处理变量引用
	if val, ok := data[expr]; ok {
		return val, nil
	}

	// 处理字符串字面量
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr[1 : len(expr)-1], nil
	}

	return expr, nil
}

// extractArgs 提取函数参数
func extractArgs(argsStr string) []string {
	argsStr = strings.TrimSuffix(argsStr, ")")
	var args []string
	depth := 0
	current := ""

	for _, ch := range argsStr {
		switch ch {
		case '(':
			depth++
			current += string(ch)
		case ')':
			depth--
			current += string(ch)
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}

	if current != "" {
		args = append(args, strings.TrimSpace(current))
	}

	return args
}

// callFunction 调用函数
func callFunction(fn interface{}, args []string, data map[string]interface{}) (interface{}, error) {
	switch f := fn.(type) {
	case func(string) string:
		if len(args) < 1 {
			return nil, fmt.Errorf("参数不足")
		}
		return f(resolveArg(args[0], data)), nil
	case func(string, string) string:
		if len(args) < 2 {
			return nil, fmt.Errorf("参数不足")
		}
		return f(resolveArg(args[0], data), resolveArg(args[1], data)), nil
	case func() string:
		return f(), nil
	case func() int64:
		return f(), nil
	default:
		return nil, fmt.Errorf("不支持的函数类型")
	}
}

// resolveArg 解析参数值
func resolveArg(arg string, data map[string]interface{}) string {
	arg = strings.TrimSpace(arg)

	// 去除引号
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
		return arg[1 : len(arg)-1]
	}

	// 从数据中获取
	if val, ok := data[arg]; ok {
		return fmt.Sprintf("%v", val)
	}

	return arg
}

// 内置函数实现

func jsonPath(jsonStr string) string {
	// 简化的 JSON Path 实现
	return jsonStr
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func base64Decode(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(data)
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func timestamp() int64 {
	return time.Now().UnixMilli()
}

func generateUUID() string {
	return uuid.New().String()
}

func regexMatch(pattern, s string) string {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return "false"
	}
	if matched {
		return "true"
	}
	return "false"
}

func regexReplace(pattern, replacement, s string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}

func urlEncode(s string) string {
	return url.QueryEscape(s)
}

func urlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}
