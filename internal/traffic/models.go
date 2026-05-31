package traffic

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Transaction 一次 HTTP 事务
type Transaction struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	DurationMs int64     `json:"duration_ms"`

	// 请求信息
	Method string `json:"method"`
	URL    string `json:"url"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Scheme string `json:"scheme"`
	Port   string `json:"port"`

	// 请求数据
	Request *RequestData `json:"request"`

	// 响应数据
	Response *ResponseData `json:"response"`

	// 网络信息
	ClientAddr string `json:"client_addr"`
	ServerIP   string `json:"server_ip"`

	// 元数据
	Bookmarked bool     `json:"bookmarked"`
	Color      string   `json:"color"`
	Notes      string   `json:"notes"`
	Tags       []string `json:"tags"`
}

// RequestData 请求数据
type RequestData struct {
	Headers     http.Header `json:"headers"`
	Body        []byte      `json:"body"`
	BodySize    int64       `json:"body_size"`
	ContentType string      `json:"content_type"`
	Raw         []byte      `json:"raw"`
}

// ResponseData 响应数据
type ResponseData struct {
	StatusCode  int            `json:"status_code"`
	StatusText  string         `json:"status_text"`
	Headers     http.Header    `json:"headers"`
	Body        []byte         `json:"body"`
	BodySize    int64          `json:"body_size"`
	ContentType string         `json:"content_type"`
	Raw         []byte         `json:"raw"`
}

// Filter 流量过滤器
type Filter struct {
	Method      []string   `json:"method,omitempty"`
	Host        []string   `json:"host,omitempty"`
	Path        string     `json:"path,omitempty"`
	StatusCode  []int      `json:"status_code,omitempty"`
	ContentType []string   `json:"content_type,omitempty"`
	MinDuration int64      `json:"min_duration,omitempty"`
	MaxDuration int64      `json:"max_duration,omitempty"`
	TimeRange   *TimeRange `json:"time_range,omitempty"`
	Search      string     `json:"search,omitempty"`
	Bookmarked  *bool      `json:"bookmarked,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// BuildSQL 构建 SQL 查询条件
func (f *Filter) BuildSQL() (string, []interface{}) {
	conditions := []string{}
	args := []interface{}{}

	if len(f.Method) > 0 {
		placeholders := make([]string, len(f.Method))
		for i, m := range f.Method {
			placeholders[i] = "?"
			args = append(args, m)
		}
		conditions = append(conditions, fmt.Sprintf("method IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(f.Host) > 0 {
		placeholders := make([]string, len(f.Host))
		for i, h := range f.Host {
			placeholders[i] = "?"
			args = append(args, h)
		}
		conditions = append(conditions, fmt.Sprintf("host IN (%s)", strings.Join(placeholders, ",")))
	}

	if f.Path != "" {
		conditions = append(conditions, "path LIKE ?")
		args = append(args, "%"+f.Path+"%")
	}

	if len(f.StatusCode) > 0 {
		placeholders := make([]string, len(f.StatusCode))
		for i, s := range f.StatusCode {
			placeholders[i] = "?"
			args = append(args, s)
		}
		conditions = append(conditions, fmt.Sprintf("status_code IN (%s)", strings.Join(placeholders, ",")))
	}

	if f.MinDuration > 0 {
		conditions = append(conditions, "duration_ms >= ?")
		args = append(args, f.MinDuration)
	}

	if f.MaxDuration > 0 {
		conditions = append(conditions, "duration_ms <= ?")
		args = append(args, f.MaxDuration)
	}

	if f.TimeRange != nil {
		if !f.TimeRange.Start.IsZero() {
			conditions = append(conditions, "timestamp >= ?")
			args = append(args, f.TimeRange.Start)
		}
		if !f.TimeRange.End.IsZero() {
			conditions = append(conditions, "timestamp <= ?")
			args = append(args, f.TimeRange.End)
		}
	}

	if f.Search != "" {
		conditions = append(conditions, "(url LIKE ? OR host LIKE ? OR notes LIKE ?)")
		search := "%" + f.Search + "%"
		args = append(args, search, search, search)
	}

	if f.Bookmarked != nil {
		conditions = append(conditions, "bookmarked = ?")
		args = append(args, *f.Bookmarked)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	return where, args
}

// TrafficStats 流量统计
type TrafficStats struct {
	TotalRequests  int64   `json:"total_requests"`
	TotalResponses int64   `json:"total_responses"`
	AvgDuration    float64 `json:"avg_duration_ms"`
	MaxDuration    int64   `json:"max_duration_ms"`
	MinDuration    int64   `json:"min_duration_ms"`
	ErrorCount     int64   `json:"error_count"`
	SuccessCount   int64   `json:"success_count"`
	HostStats      []HostStat `json:"host_stats"`
	MethodStats    []MethodStat `json:"method_stats"`
	StatusStats    []StatusStat `json:"status_stats"`
}

// HostStat 主机统计
type HostStat struct {
	Host     string `json:"host"`
	Count    int64  `json:"count"`
	AvgTime  float64 `json:"avg_time_ms"`
}

// MethodStat 方法统计
type MethodStat struct {
	Method string `json:"method"`
	Count  int64  `json:"count"`
}

// StatusStat 状态码统计
type StatusStat struct {
	StatusCode int   `json:"status_code"`
	Count      int64 `json:"count"`
}
