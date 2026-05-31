package rules

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Engine 规则引擎
type Engine struct {
	store    *Store
	matcher  *Matcher
	executor *ActionExecutor
	mu       sync.RWMutex
	rules    []*Rule
	stats    map[string]int
}

// NewEngine 创建新的规则引擎
func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		store:    NewStore(db),
		matcher:  NewMatcher(),
		executor: NewActionExecutor(),
		stats:    make(map[string]int),
	}
}

// Init 初始化规则引擎
func (e *Engine) Init() error {
	if err := e.store.Init(); err != nil {
		return err
	}

	// 加载规则到内存
	return e.reload()
}

// reload 重新加载规则
func (e *Engine) reload() error {
	rules, err := e.store.GetEnabled()
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.rules = rules
	e.mu.Unlock()

	log.Printf("[INFO] 规则引擎加载 %d 条规则", len(rules))
	return nil
}

// ProcessRequest 处理请求
func (e *Engine) ProcessRequest(req *http.Request) (*http.Request, *http.Response, error) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	// 匹配规则
	matched := e.matcher.MatchRules(rules, req.Method, req.URL.String(), req.Host, req.Header)
	if len(matched) == 0 {
		return req, nil, nil
	}

	// 按优先级排序
	sorted := SortByPriority(matched)

	// 执行第一个匹配的规则
	rule := sorted[0]
	e.stats[rule.ID]++

	log.Printf("[RULE] 规则命中: %s (%s)", rule.Name, rule.ID)

	// 执行动作
	return e.executor.Execute(rule, req, nil)
}

// ProcessResponse 处理响应
func (e *Engine) ProcessResponse(req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	// 匹配规则
	matched := e.matcher.MatchRules(rules, req.Method, req.URL.String(), req.Host, req.Header)
	if len(matched) == 0 {
		return req, resp, nil
	}

	// 按优先级排序
	sorted := SortByPriority(matched)

	// 执行第一个匹配的规则
	rule := sorted[0]

	// 只有修改响应和阻止动作才在响应阶段执行
	if rule.Action.Type == ActionModifyResponse || rule.Action.Type == ActionBlock {
		e.stats[rule.ID]++
		log.Printf("[RULE] 规则命中(响应): %s (%s)", rule.Name, rule.ID)
		return e.executor.Execute(rule, req, resp)
	}

	return req, resp, nil
}

// CreateRule 创建规则
func (e *Engine) CreateRule(rule *Rule) error {
	if err := e.store.Create(rule); err != nil {
		return err
	}
	return e.reload()
}

// GetRule 获取规则
func (e *Engine) GetRule(id string) (*Rule, error) {
	return e.store.GetByID(id)
}

// ListRules 获取规则列表
func (e *Engine) ListRules() ([]*Rule, error) {
	return e.store.List()
}

// UpdateRule 更新规则
func (e *Engine) UpdateRule(rule *Rule) error {
	if err := e.store.Update(rule); err != nil {
		return err
	}
	return e.reload()
}

// DeleteRule 删除规则
func (e *Engine) DeleteRule(id string) error {
	if err := e.store.Delete(id); err != nil {
		return err
	}
	return e.reload()
}

// ToggleRule 切换规则状态
func (e *Engine) ToggleRule(id string) (bool, error) {
	enabled, err := e.store.Toggle(id)
	if err != nil {
		return false, err
	}
	if err := e.reload(); err != nil {
		return false, err
	}
	return enabled, nil
}

// ReorderRules 重新排序规则
func (e *Engine) ReorderRules(ids []string) error {
	if err := e.store.Reorder(ids); err != nil {
		return err
	}
	return e.reload()
}

// GetStats 获取规则统计
func (e *Engine) GetStats() (*RuleStats, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := &RuleStats{
		HitCounts: make(map[string]int),
	}

	// 复制统计信息
	for k, v := range e.stats {
		stats.HitCounts[k] = v
	}

	// 获取规则数量
	rules, err := e.store.List()
	if err != nil {
		return nil, err
	}

	stats.TotalRules = len(rules)
	for _, r := range rules {
		if r.Enabled {
			stats.EnabledRules++
		} else {
			stats.DisabledRules++
		}
	}

	return stats, nil
}

// ImportRules 导入规则
func (e *Engine) ImportRules(data []byte) (int, error) {
	var rules []*Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return 0, err
	}

	count := 0
	for _, rule := range rules {
		if err := e.store.Create(rule); err != nil {
			log.Printf("[WARN] 导入规则失败: %v", err)
			continue
		}
		count++
	}

	if err := e.reload(); err != nil {
		return count, err
	}

	return count, nil
}

// ExportRules 导出规则
func (e *Engine) ExportRules() ([]byte, error) {
	rules, err := e.store.List()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(rules, "", "  ")
}
