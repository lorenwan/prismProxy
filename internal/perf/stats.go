package perf

import (
	"sort"
	"sync"
	"time"
)

// StatsCollector 统计收集器
type StatsCollector struct {
	mu             sync.RWMutex
	durations      []int64
	window         *SlidingWindow
	lastCleanup    time.Time
	cleanupInterval time.Duration
}

// NewStatsCollector 创建统计收集器
func NewStatsCollector(windowSize int) *StatsCollector {
	return &StatsCollector{
		durations:       make([]int64, 0),
		window:          NewSlidingWindow(windowSize),
		lastCleanup:     time.Now(),
		cleanupInterval: 5 * time.Minute,
	}
}

// Record 记录一次请求
func (c *StatsCollector) Record(durationMs int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.durations = append(c.durations, durationMs)
	c.window.Add(durationMs)

	// 定期清理旧数据
	if time.Since(c.lastCleanup) > c.cleanupInterval {
		c.cleanup()
	}
}

// cleanup 清理旧数据
func (c *StatsCollector) cleanup() {
	// 保留最近 10000 条记录
	maxRecords := 10000
	if len(c.durations) > maxRecords {
		c.durations = c.durations[len(c.durations)-maxRecords:]
	}
	c.lastCleanup = time.Now()
}

// GetPercentiles 获取百分位
func (c *StatsCollector) GetPercentiles() *PercentileResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.durations) == 0 {
		return &PercentileResult{}
	}

	// 复制并排序
	sorted := make([]int64, len(c.durations))
	copy(sorted, c.durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	return CalculatePercentiles(sorted)
}

// GetWindowAverage 获取滑动窗口平均值
func (c *StatsCollector) GetWindowAverage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.window.GetAverage()
}

// GetTotal 获取总请求数
func (c *StatsCollector) GetTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return int64(len(c.durations))
}

// CalculatePercentiles 计算百分位
func CalculatePercentiles(sorted []int64) *PercentileResult {
	if len(sorted) == 0 {
		return &PercentileResult{}
	}

	return &PercentileResult{
		P50: percentile(sorted, 50),
		P90: percentile(sorted, 90),
		P95: percentile(sorted, 95),
		P99: percentile(sorted, 99),
	}
}

// percentile 计算指定百分位
func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	index := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	// 线性插值
	weight := index - float64(lower)
	return int64(float64(sorted[lower])*(1-weight) + float64(sorted[upper])*weight)
}

// CalculateStdDev 计算标准差
func CalculateStdDev(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	// 计算平均值
	var sum int64
	for _, v := range values {
		sum += v
	}
	mean := float64(sum) / float64(len(values))

	// 计算方差
	var variance float64
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	// 返回标准差
	return sqrt(variance)
}

// sqrt 平方根（简单实现）
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}

	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// CalculateMedian 计算中位数
func CalculateMedian(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]int64, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return float64(sorted[mid-1]+sorted[mid]) / 2
	}
	return float64(sorted[mid])
}

// CalculateMean 计算平均值
func CalculateMean(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum int64
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// GetMin 获取最小值
func GetMin(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// GetMax 获取最大值
func GetMax(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// GetSum 获取总和
func GetSum(values []int64) int64 {
	var sum int64
	for _, v := range values {
		sum += v
	}
	return sum
}

// BucketDurations 将持续时间分桶
func BucketDurations(durations []int64, bucketSize int64) map[int64]int64 {
	buckets := make(map[int64]int64)

	for _, d := range durations {
		bucket := (d / bucketSize) * bucketSize
		buckets[bucket]++
	}

	return buckets
}

// GetDurationDistribution 获取持续时间分布
func GetDurationDistribution(durations []int64) map[string]int64 {
	dist := map[string]int64{
		"<100ms":    0,
		"100-500ms": 0,
		"500ms-1s":  0,
		"1-3s":      0,
		"3-10s":     0,
		">10s":      0,
	}

	for _, d := range durations {
		switch {
		case d < 100:
			dist["<100ms"]++
		case d < 500:
			dist["100-500ms"]++
		case d < 1000:
			dist["500ms-1s"]++
		case d < 3000:
			dist["1-3s"]++
		case d < 10000:
			dist["3-10s"]++
		default:
			dist[">10s"]++
		}
	}

	return dist
}
