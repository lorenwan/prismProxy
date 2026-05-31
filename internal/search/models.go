package search

import (
	"fmt"
	"strings"
	"time"

	"prismproxy/internal/traffic"
)

// SearchQuery 搜索查询
type SearchQuery struct {
	Query    string   `json:"query"`
	Filters  []Filter `json:"filters,omitempty"`
	Sort     string   `json:"sort,omitempty"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// Filter 过滤条件
type Filter struct {
	Field    string      `json:"field"`
	Operator FilterOp    `json:"operator"`
	Value    interface{} `json:"value"`
}

// FilterOp 过滤操作符
type FilterOp string

const (
	FilterOpEq       FilterOp = "eq"
	FilterOpNe       FilterOp = "ne"
	FilterOpGt       FilterOp = "gt"
	FilterOpLt       FilterOp = "lt"
	FilterOpContains FilterOp = "contains"
	FilterOpRegex    FilterOp = "regex"
)

// SavedFilter 已保存的过滤器
type SavedFilter struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Query     string    `json:"query"`
	Filters   []Filter  `json:"filters,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Items      []*traffic.Transaction `json:"items"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// FilterField 过滤字段
type FilterField string

const (
	FieldMethod     FilterField = "method"
	FieldHost       FilterField = "host"
	FieldPath       FilterField = "path"
	FieldStatusCode FilterField = "status_code"
	FieldDuration   FilterField = "duration"
	FieldContentType FilterField = "content_type"
	FieldTimestamp  FilterField = "timestamp"
	FieldBookmarked FilterField = "bookmarked"
	FieldTags       FilterField = "tags"
)

// BuildWhereClause 构建 WHERE 子句
func (q *SearchQuery) BuildWhereClause() (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 全文搜索条件
	if q.Query != "" {
		conditions = append(conditions, "(url LIKE ? OR host LIKE ? OR path LIKE ?)")
		searchPattern := "%" + q.Query + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	// 过滤条件
	for _, filter := range q.Filters {
		clause, filterArgs := buildFilterClause(filter)
		if clause != "" {
			conditions = append(conditions, clause)
			args = append(args, filterArgs...)
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// buildFilterClause 构建过滤子句
func buildFilterClause(filter Filter) (string, []interface{}) {
	field := string(filter.Field)
	value := filter.Value

	switch filter.Operator {
	case FilterOpEq:
		return fmt.Sprintf("%s = ?", field), []interface{}{value}
	case FilterOpNe:
		return fmt.Sprintf("%s != ?", field), []interface{}{value}
	case FilterOpGt:
		return fmt.Sprintf("%s > ?", field), []interface{}{value}
	case FilterOpLt:
		return fmt.Sprintf("%s < ?", field), []interface{}{value}
	case FilterOpContains:
		return fmt.Sprintf("%s LIKE ?", field), []interface{}{"%" + fmt.Sprintf("%v", value) + "%"}
	case FilterOpRegex:
		// SQLite 不支持原生正则，使用 LIKE 模拟
		return fmt.Sprintf("%s LIKE ?", field), []interface{}{"%" + fmt.Sprintf("%v", value) + "%"}
	default:
		return "", nil
	}
}

// BuildOrderBy 构建 ORDER BY 子句
func (q *SearchQuery) BuildOrderBy() string {
	if q.Sort == "" {
		return "ORDER BY timestamp DESC"
	}

	// 解析排序字段和方向
	parts := strings.Split(q.Sort, ":")
	if len(parts) != 2 {
		return "ORDER BY timestamp DESC"
	}

	field := parts[0]
	direction := strings.ToUpper(parts[1])

	// 验证字段
	validFields := map[string]bool{
		"timestamp":   true,
		"duration_ms": true,
		"method":      true,
		"host":        true,
		"status_code": true,
	}

	if !validFields[field] {
		return "ORDER BY timestamp DESC"
	}

	if direction != "ASC" && direction != "DESC" {
		direction = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", field, direction)
}

// GetOffset 获取偏移量
func (q *SearchQuery) GetOffset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	return (q.Page - 1) * q.PageSize
}

// GetLimit 获取限制
func (q *SearchQuery) GetLimit() int {
	if q.PageSize <= 0 {
		return 20
	}
	return q.PageSize
}
