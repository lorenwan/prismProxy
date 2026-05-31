package perf

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// PerfAnalyzer 性能分析器
type PerfAnalyzer struct {
	db *sql.DB
}

// NewAnalyzer 创建新的性能分析器
func NewAnalyzer(db *sql.DB) *PerfAnalyzer {
	return &PerfAnalyzer{db: db}
}

// GetStats 获取性能统计
func (a *PerfAnalyzer) GetStats(since time.Time) (*PerfStats, error) {
	stats := &PerfStats{
		TimeRange: TimeRange{
			Start: since,
			End:   time.Now(),
		},
	}

	// 查询基本统计
	query := `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(duration_ms), 0) as total_duration,
			COALESCE(MIN(duration_ms), 0) as min_duration,
			COALESCE(MAX(duration_ms), 0) as max_duration
		FROM traffic
		WHERE timestamp >= ?
	`

	err := a.db.QueryRow(query, since).Scan(
		&stats.TotalRequests,
		&stats.TotalDuration,
		&stats.MinDuration,
		&stats.MaxDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("查询统计失败: %w", err)
	}

	// 计算平均值
	if stats.TotalRequests > 0 {
		stats.AvgDuration = float64(stats.TotalDuration) / float64(stats.TotalRequests)
	}

	// 查询慢请求数量（>1000ms）
	slowQuery := `
		SELECT COUNT(*) FROM traffic
		WHERE timestamp >= ? AND duration_ms > 1000
	`
	a.db.QueryRow(slowQuery, since).Scan(&stats.SlowRequests)

	// 计算百分位
	percentiles, err := a.getPercentiles(since)
	if err == nil {
		stats.P50 = percentiles.P50
		stats.P90 = percentiles.P90
		stats.P99 = percentiles.P99
	}

	return stats, nil
}

// getPercentiles 计算百分位
func (a *PerfAnalyzer) getPercentiles(since time.Time) (*PercentileResult, error) {
	query := `
		SELECT duration_ms
		FROM traffic
		WHERE timestamp >= ?
		ORDER BY duration_ms ASC
	`

	rows, err := a.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var durations []int64
	for rows.Next() {
		var d int64
		if err := rows.Scan(&d); err != nil {
			continue
		}
		durations = append(durations, d)
	}

	return CalculatePercentiles(durations), nil
}

// GetSlowRequests 获取慢请求
func (a *PerfAnalyzer) GetSlowRequests(threshold time.Duration, limit int) ([]SlowRequest, error) {
	query := `
		SELECT id, url, method, host, status_code, duration_ms, timestamp
		FROM traffic
		WHERE duration_ms > ?
		ORDER BY duration_ms DESC
		LIMIT ?
	`

	rows, err := a.db.Query(query, threshold.Milliseconds(), limit)
	if err != nil {
		return nil, fmt.Errorf("查询慢请求失败: %w", err)
	}
	defer rows.Close()

	var requests []SlowRequest
	for rows.Next() {
		req := SlowRequest{}
		err := rows.Scan(
			&req.TransactionID,
			&req.URL,
			&req.Method,
			&req.Host,
			&req.StatusCode,
			&req.Duration,
			&req.Timestamp,
		)
		if err != nil {
			continue
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetDomainStats 获取域名统计
func (a *PerfAnalyzer) GetDomainStats() ([]DomainStats, error) {
	query := `
		SELECT
			host,
			COUNT(*) as request_count,
			AVG(duration_ms) as avg_duration,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count,
			SUM(duration_ms) as total_duration
		FROM traffic
		GROUP BY host
		ORDER BY request_count DESC
	`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询域名统计失败: %w", err)
	}
	defer rows.Close()

	var stats []DomainStats
	for rows.Next() {
		stat := DomainStats{}
		err := rows.Scan(
			&stat.Domain,
			&stat.RequestCount,
			&stat.AvgDuration,
			&stat.ErrorCount,
			&stat.TotalDuration,
		)
		if err != nil {
			continue
		}

		// 计算错误率
		if stat.RequestCount > 0 {
			stat.ErrorRate = float64(stat.ErrorCount) / float64(stat.RequestCount) * 100
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTimelineStats 获取时间线统计
func (a *PerfAnalyzer) GetTimelineStats(since time.Time, interval time.Duration) ([]TimelinePoint, error) {
	// 根据间隔选择时间格式
	var timeFormat string
	switch {
	case interval <= time.Minute:
		timeFormat = "%Y-%m-%d %H:%M:00"
	case interval <= time.Hour:
		timeFormat = "%Y-%m-%d %H:00:00"
	case interval <= 24*time.Hour:
		timeFormat = "%Y-%m-%d 00:00:00"
	default:
		timeFormat = "%Y-%m-01 00:00:00"
	}

	query := fmt.Sprintf(`
		SELECT
			MAX(CASE WHEN strftime('%s', timestamp) = strftime('%s', MAX(timestamp)) THEN timestamp END) as ts,
			COUNT(*) as request_count,
			AVG(duration_ms) as avg_duration,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count
		FROM traffic
		WHERE timestamp >= ?
		GROUP BY strftime('%s', timestamp)
		ORDER BY ts ASC
	`, timeFormat, timeFormat, timeFormat)

	rows, err := a.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("查询时间线失败: %w", err)
	}
	defer rows.Close()

	var points []TimelinePoint
	for rows.Next() {
		point := TimelinePoint{}
		err := rows.Scan(
			&point.Timestamp,
			&point.RequestCount,
			&point.AvgDuration,
			&point.ErrorCount,
		)
		if err != nil {
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

// RecordTransaction 记录事务（用于实时统计）
func (a *PerfAnalyzer) RecordTransaction(durationMs int64, statusCode int, host string) {
	// 这里可以实现内存中的实时统计
	// 暂时只记录到数据库，由 GetStats 等方法查询
}

// GetMethodStats 获取请求方法统计
func (a *PerfAnalyzer) GetMethodStats(since time.Time) ([]MethodStats, error) {
	query := `
		SELECT method, COUNT(*) as count
		FROM traffic
		WHERE timestamp >= ?
		GROUP BY method
		ORDER BY count DESC
	`

	rows, err := a.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("查询方法统计失败: %w", err)
	}
	defer rows.Close()

	var stats []MethodStats
	for rows.Next() {
		stat := MethodStats{}
		if err := rows.Scan(&stat.Method, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetStatusCodeStats 获取状态码统计
func (a *PerfAnalyzer) GetStatusCodeStats(since time.Time) ([]StatusCodeStats, error) {
	query := `
		SELECT status_code, COUNT(*) as count
		FROM traffic
		WHERE timestamp >= ?
		GROUP BY status_code
		ORDER BY count DESC
	`

	rows, err := a.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("查询状态码统计失败: %w", err)
	}
	defer rows.Close()

	var stats []StatusCodeStats
	for rows.Next() {
		stat := StatusCodeStats{}
		if err := rows.Scan(&stat.StatusCode, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetRecentStats 获取最近 N 分钟的统计
func (a *PerfAnalyzer) GetRecentStats(minutes int) (*PerfStats, error) {
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	return a.GetStats(since)
}

// SortSlowRequests 排序慢请求
func SortSlowRequests(requests []SlowRequest, sortBy string, ascending bool) {
	sort.Slice(requests, func(i, j int) bool {
		switch sortBy {
		case "duration":
			if ascending {
				return requests[i].Duration < requests[j].Duration
			}
			return requests[i].Duration > requests[j].Duration
		case "timestamp":
			if ascending {
				return requests[i].Timestamp.Before(requests[j].Timestamp)
			}
			return requests[i].Timestamp.After(requests[j].Timestamp)
		default:
			return requests[i].Duration > requests[j].Duration
		}
	})
}
