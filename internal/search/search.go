package search

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"prismproxy/internal/traffic"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	db *sql.DB
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(db *sql.DB) *SearchEngine {
	return &SearchEngine{db: db}
}

// FullTextSearch 全文搜索
func (e *SearchEngine) FullTextSearch(query string, page, pageSize int) (*SearchResult, error) {
	q := &SearchQuery{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
	}
	return e.Search(q)
}

// Search 执行搜索
func (e *SearchEngine) Search(query *SearchQuery) (*SearchResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// 构建查询
	whereClause, args := query.BuildWhereClause()
	orderBy := query.BuildOrderBy()

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traffic %s", whereClause)
	var total int
	if err := e.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, method, url, host, path, scheme, port,
			status_code, duration_ms, req_headers, req_body,
			resp_headers, resp_body, timestamp
		FROM traffic %s %s LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	queryArgs := append(args, query.GetLimit(), query.GetOffset())
	rows, err := e.db.Query(dataQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	items := make([]*traffic.Transaction, 0)
	for rows.Next() {
		t := &traffic.Transaction{
			Request:  &traffic.RequestData{},
			Response: &traffic.ResponseData{},
		}
		var reqHeaders, respHeaders string
		var timestamp time.Time

		err := rows.Scan(
			&t.ID, &t.Method, &t.URL, &t.Host, &t.Path, &t.Scheme, &t.Port,
			&t.Response.StatusCode, &t.DurationMs, &reqHeaders, &t.Request.Body,
			&respHeaders, &t.Response.Body, &timestamp,
		)
		if err != nil {
			continue
		}

		t.Timestamp = timestamp

		// 解析 headers
		t.Request.Headers = parseHeaders(reqHeaders)
		t.Response.Headers = parseHeaders(respHeaders)

		items = append(items, t)
	}

	// 计算总页数
	totalPages := total / query.PageSize
	if total%query.PageSize > 0 {
		totalPages++
	}

	return &SearchResult{
		Items:      items,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchWithFilters 使用过滤器搜索
func (e *SearchEngine) SearchWithFilters(filters []Filter, page, pageSize int) (*SearchResult, error) {
	q := &SearchQuery{
		Filters:  filters,
		Page:     page,
		PageSize: pageSize,
	}
	return e.Search(q)
}

// SearchByMethod 按方法搜索
func (e *SearchEngine) SearchByMethod(method string, page, pageSize int) (*SearchResult, error) {
	return e.SearchWithFilters([]Filter{
		{Field: "method", Operator: FilterOpEq, Value: method},
	}, page, pageSize)
}

// SearchByHost 按主机搜索
func (e *SearchEngine) SearchByHost(host string, page, pageSize int) (*SearchResult, error) {
	return e.SearchWithFilters([]Filter{
		{Field: "host", Operator: FilterOpContains, Value: host},
	}, page, pageSize)
}

// SearchByStatusCode 按状态码搜索
func (e *SearchEngine) SearchByStatusCode(statusCode int, page, pageSize int) (*SearchResult, error) {
	return e.SearchWithFilters([]Filter{
		{Field: "status_code", Operator: FilterOpEq, Value: statusCode},
	}, page, pageSize)
}

// SearchSlowRequests 搜索慢请求
func (e *SearchEngine) SearchSlowRequests(thresholdMs int64, page, pageSize int) (*SearchResult, error) {
	return e.SearchWithFilters([]Filter{
		{Field: "duration_ms", Operator: FilterOpGt, Value: thresholdMs},
	}, page, pageSize)
}

// SearchByTimeRange 按时间范围搜索
func (e *SearchEngine) SearchByTimeRange(start, end time.Time, page, pageSize int) (*SearchResult, error) {
	return e.SearchWithFilters([]Filter{
		{Field: "timestamp", Operator: FilterOpGt, Value: start},
		{Field: "timestamp", Operator: FilterOpLt, Value: end},
	}, page, pageSize)
}

// parseHeaders 解析 headers JSON
func parseHeaders(jsonStr string) http.Header {
	headers := make(http.Header)
	if jsonStr == "" {
		return headers
	}

	var raw map[string][]string
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		// 尝试解析为 map[string]string
		var simple map[string]string
		if err := json.Unmarshal([]byte(jsonStr), &simple); err != nil {
			return headers
		}
		for k, v := range simple {
			headers.Set(k, v)
		}
		return headers
	}

	for k, v := range raw {
		headers[k] = v
	}

	return headers
}

// BuildSearchStats 构建搜索统计
func (e *SearchEngine) BuildSearchStats(query *SearchQuery) (map[string]interface{}, error) {
	whereClause, args := query.BuildWhereClause()

	stats := make(map[string]interface{})

	// 总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traffic %s", whereClause)
	var total int
	if err := e.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}
	stats["total"] = total

	// 方法分布
	methodQuery := fmt.Sprintf(`
		SELECT method, COUNT(*) as count
		FROM traffic %s
		GROUP BY method ORDER BY count DESC
	`, whereClause)
	rows, err := e.db.Query(methodQuery, args...)
	if err == nil {
		defer rows.Close()
		methods := make(map[string]int)
		for rows.Next() {
			var method string
			var count int
			if rows.Scan(&method, &count) == nil {
				methods[method] = count
			}
		}
		stats["methods"] = methods
	}

	// 状态码分布
	statusQuery := fmt.Sprintf(`
		SELECT status_code, COUNT(*) as count
		FROM traffic %s
		GROUP BY status_code ORDER BY count DESC
		LIMIT 10
	`, whereClause)
	rows, err = e.db.Query(statusQuery, args...)
	if err == nil {
		defer rows.Close()
		statuses := make(map[int]int)
		for rows.Next() {
			var status, count int
			if rows.Scan(&status, &count) == nil {
				statuses[status] = count
			}
		}
		stats["status_codes"] = statuses
	}

	return stats, nil
}
