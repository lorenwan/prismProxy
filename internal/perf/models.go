package perf

import "time"

// PerfStats 性能统计
type PerfStats struct {
	TotalRequests int64         `json:"total_requests"`
	AvgDuration   float64       `json:"avg_duration_ms"`
	P50           int64         `json:"p50_ms"`
	P90           int64         `json:"p90_ms"`
	P99           int64         `json:"p99_ms"`
	SlowRequests  int64         `json:"slow_requests"`
	MinDuration   int64         `json:"min_duration_ms"`
	MaxDuration   int64         `json:"max_duration_ms"`
	TotalDuration int64         `json:"total_duration_ms"`
	TimeRange     TimeRange     `json:"time_range"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SlowRequest 慢请求
type SlowRequest struct {
	TransactionID int64     `json:"transaction_id"`
	URL           string    `json:"url"`
	Method        string    `json:"method"`
	Host          string    `json:"host"`
	StatusCode    int       `json:"status_code"`
	Duration      int64     `json:"duration_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// DomainStats 域名统计
type DomainStats struct {
	Domain        string  `json:"domain"`
	RequestCount  int64   `json:"request_count"`
	AvgDuration   float64 `json:"avg_duration_ms"`
	ErrorCount    int64   `json:"error_count"`
	ErrorRate     float64 `json:"error_rate"`
	TotalDuration int64   `json:"total_duration_ms"`
}

// TimelinePoint 时间线数据点
type TimelinePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int64     `json:"request_count"`
	AvgDuration  float64   `json:"avg_duration_ms"`
	ErrorCount   int64     `json:"error_count"`
}

// StatusCodeStats 状态码统计
type StatusCodeStats struct {
	StatusCode int   `json:"status_code"`
	Count      int64 `json:"count"`
}

// MethodStats 请求方法统计
type MethodStats struct {
	Method string `json:"method"`
	Count  int64  `json:"count"`
}

// PercentileResult 百分位结果
type PercentileResult struct {
	P50 int64 `json:"p50"`
	P90 int64 `json:"p90"`
	P95 int64 `json:"p95"`
	P99 int64 `json:"p99"`
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	Size     int
	Data     []int64
	Position int
	Full     bool
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(size int) *SlidingWindow {
	return &SlidingWindow{
		Size: size,
		Data: make([]int64, size),
	}
}

// Add 添加数据
func (w *SlidingWindow) Add(value int64) {
	w.Data[w.Position] = value
	w.Position = (w.Position + 1) % w.Size
	if w.Position == 0 {
		w.Full = true
	}
}

// GetValues 获取所有值
func (w *SlidingWindow) GetValues() []int64 {
	if w.Full {
		return w.Data
	}
	return w.Data[:w.Position]
}

// GetAverage 获取平均值
func (w *SlidingWindow) GetAverage() float64 {
	values := w.GetValues()
	if len(values) == 0 {
		return 0
	}

	var sum int64
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}
