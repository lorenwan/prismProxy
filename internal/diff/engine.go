package diff

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// DiffEngine 对比引擎
type DiffEngine struct{}

// NewEngine 创建新的对比引擎
func NewEngine() *DiffEngine {
	return &DiffEngine{}
}

// CompareHeaders 对比 Headers
func (e *DiffEngine) CompareHeaders(left, right map[string][]string) DiffResult {
	result := DiffResult{
		Type:    DiffTypeHeaders,
		Entries: make([]DiffEntry, 0),
	}

	// 收集所有 key
	allKeys := make(map[string]bool)
	for k := range left {
		allKeys[strings.ToLower(k)] = true
	}
	for k := range right {
		allKeys[strings.ToLower(k)] = true
	}

	for key := range allKeys {
		leftVal, leftOk := findHeader(left, key)
		rightVal, rightOk := findHeader(right, key)

		entry := DiffEntry{
			Path: key,
		}

		switch {
		case leftOk && !rightOk:
			entry.Left = leftVal
			entry.Status = StatusRemoved
		case !leftOk && rightOk:
			entry.Right = rightVal
			entry.Status = StatusAdded
		case leftVal != rightVal:
			entry.Left = leftVal
			entry.Right = rightVal
			entry.Status = StatusModified
		default:
			entry.Left = leftVal
			entry.Status = StatusUnchanged
		}

		result.Entries = append(result.Entries, entry)
	}

	// 排序
	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Path < result.Entries[j].Path
	})

	return result
}

// findHeader 在 header map 中查找（忽略大小写）
func findHeader(headers map[string][]string, key string) (string, bool) {
	for k, v := range headers {
		if strings.ToLower(k) == key {
			return strings.Join(v, ", "), true
		}
	}
	return "", false
}

// CompareBody 对比 Body
func (e *DiffEngine) CompareBody(left, right []byte) DiffResult {
	result := DiffResult{
		Type:    DiffTypeBody,
		Entries: make([]DiffEntry, 0),
	}

	leftStr := string(left)
	rightStr := string(right)

	// 逐行对比
	leftLines := strings.Split(leftStr, "\n")
	rightLines := strings.Split(rightStr, "\n")

	// 使用 patience diff 算法
	diff := patienceDiff(leftLines, rightLines)

	for _, d := range diff {
		entry := DiffEntry{
			Path:   d.Path,
			Left:   d.Left,
			Right:  d.Right,
			Status: d.Status,
		}
		result.Entries = append(result.Entries, entry)
	}

	return result
}

// diffLine diff 行
type diffLine struct {
	Path   string
	Left   string
	Right  string
	Status DiffStatus
}

// patienceDiff patience diff 算法实现
func patienceDiff(left, right []string) []diffLine {
	var result []diffLine

	// 简化的 diff 实现：使用 LCS 算法
	lcs := longestCommonSubsequence(left, right)

	leftIdx, rightIdx, lcsIdx := 0, 0, 0

	for leftIdx < len(left) || rightIdx < len(right) {
		if lcsIdx < len(lcs) {
			// 处理 LCS 之前的行
			for leftIdx < lcs[lcsIdx].Left && rightIdx < lcs[lcsIdx].Right {
				if leftIdx < lcs[lcsIdx].Left && rightIdx < lcs[lcsIdx].Right {
					// 两边都修改了
					result = append(result, diffLine{
						Path:   fmt.Sprintf("line:%d", leftIdx+1),
						Left:   left[leftIdx],
						Right:  right[rightIdx],
						Status: StatusModified,
					})
					leftIdx++
					rightIdx++
				}
			}

			for leftIdx < lcs[lcsIdx].Left {
				result = append(result, diffLine{
					Path:   fmt.Sprintf("line:%d", leftIdx+1),
					Left:   left[leftIdx],
					Status: StatusRemoved,
				})
				leftIdx++
			}

			for rightIdx < lcs[lcsIdx].Right {
				result = append(result, diffLine{
					Path:   fmt.Sprintf("line:%d", rightIdx+1),
					Right:  right[rightIdx],
					Status: StatusAdded,
				})
				rightIdx++
			}

			// LCS 行
			result = append(result, diffLine{
				Path:   fmt.Sprintf("line:%d", leftIdx+1),
				Left:   left[leftIdx],
				Right:  right[rightIdx],
				Status: StatusUnchanged,
			})
			leftIdx++
			rightIdx++
			lcsIdx++
		} else {
			// 处理剩余行
			if leftIdx < len(left) && rightIdx < len(right) {
				result = append(result, diffLine{
					Path:   fmt.Sprintf("line:%d", leftIdx+1),
					Left:   left[leftIdx],
					Right:  right[rightIdx],
					Status: StatusModified,
				})
				leftIdx++
				rightIdx++
			} else if leftIdx < len(left) {
				result = append(result, diffLine{
					Path:   fmt.Sprintf("line:%d", leftIdx+1),
					Left:   left[leftIdx],
					Status: StatusRemoved,
				})
				leftIdx++
			} else {
				result = append(result, diffLine{
					Path:   fmt.Sprintf("line:%d", rightIdx+1),
					Right:  right[rightIdx],
					Status: StatusAdded,
				})
				rightIdx++
			}
		}
	}

	return result
}

// lcsPair LCS 配对
type lcsPair struct {
	Left  int
	Right int
}

// longestCommonSubsequence 最长公共子序列
func longestCommonSubsequence(left, right []string) []lcsPair {
	m, n := len(left), len(right)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if left[i-1] == right[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// 回溯找出 LCS
	var result []lcsPair
	i, j := m, n
	for i > 0 && j > 0 {
		if left[i-1] == right[j-1] {
			result = append([]lcsPair{{Left: i - 1, Right: j - 1}}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return result
}

// CompareJSON 对比 JSON
func (e *DiffEngine) CompareJSON(left, right []byte) JSONDiffResult {
	result := JSONDiffResult{
		Diffs: make([]JSONDiffEntry, 0),
	}

	var leftObj, rightObj interface{}
	json.Unmarshal(left, &leftObj)
	json.Unmarshal(right, &rightObj)

	// 递归对比
	e.compareRecursive("", leftObj, rightObj, &result)

	// 计算摘要
	for _, d := range result.Diffs {
		result.Summary.TotalFields++
		switch d.Status {
		case StatusAdded:
			result.Summary.Added++
		case StatusRemoved:
			result.Summary.Removed++
		case StatusModified:
			result.Summary.Modified++
		case StatusUnchanged:
			result.Summary.Unchanged++
		}
	}

	return result
}

// compareRecursive 递归对比
func (e *DiffEngine) compareRecursive(path string, left, right interface{}, result *JSONDiffResult) {
	// 类型不同
	if getType(left) != getType(right) {
		result.Diffs = append(result.Diffs, JSONDiffEntry{
			Path:   path,
			Left:   left,
			Right:  right,
			Status: StatusModified,
			Type:   "mixed",
		})
		return
	}

	switch l := left.(type) {
	case map[string]interface{}:
		r, _ := right.(map[string]interface{})
		e.compareObjects(path, l, r, result)
	case []interface{}:
		r, _ := right.([]interface{})
		e.compareArrays(path, l, r, result)
	default:
		// 基本类型
		if left == right {
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   path,
				Left:   left,
				Right:  right,
				Status: StatusUnchanged,
				Type:   getType(left),
			})
		} else {
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   path,
				Left:   left,
				Right:  right,
				Status: StatusModified,
				Type:   getType(left),
			})
		}
	}
}

// compareObjects 对比对象
func (e *DiffEngine) compareObjects(path string, left, right map[string]interface{}, result *JSONDiffResult) {
	// 收集所有 key
	allKeys := make(map[string]bool)
	for k := range left {
		allKeys[k] = true
	}
	for k := range right {
		allKeys[k] = true
	}

	for key := range allKeys {
		childPath := key
		if path != "" {
			childPath = path + "." + key
		}

		leftVal, leftOk := left[key]
		rightVal, rightOk := right[key]

		switch {
		case leftOk && !rightOk:
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   childPath,
				Left:   leftVal,
				Status: StatusRemoved,
				Type:   getType(leftVal),
			})
		case !leftOk && rightOk:
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   childPath,
				Right:  rightVal,
				Status: StatusAdded,
				Type:   getType(rightVal),
			})
		default:
			e.compareRecursive(childPath, leftVal, rightVal, result)
		}
	}
}

// compareArrays 对比数组
func (e *DiffEngine) compareArrays(path string, left, right []interface{}, result *JSONDiffResult) {
	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}

	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s[%d]", path, i)

		switch {
		case i >= len(left):
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   childPath,
				Right:  right[i],
				Status: StatusAdded,
				Type:   getType(right[i]),
			})
		case i >= len(right):
			result.Diffs = append(result.Diffs, JSONDiffEntry{
				Path:   childPath,
				Left:   left[i],
				Status: StatusRemoved,
				Type:   getType(left[i]),
			})
		default:
			e.compareRecursive(childPath, left[i], right[i], result)
		}
	}
}

// getType 获取类型
func getType(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// CompareQuery 对比 Query 参数
func (e *DiffEngine) CompareQuery(left, right url.Values) DiffResult {
	result := DiffResult{
		Type:    DiffTypeQuery,
		Entries: make([]DiffEntry, 0),
	}

	// 收集所有 key
	allKeys := make(map[string]bool)
	for k := range left {
		allKeys[k] = true
	}
	for k := range right {
		allKeys[k] = true
	}

	for key := range allKeys {
		leftVals := left[key]
		rightVals := right[key]

		entry := DiffEntry{
			Path: key,
		}

		leftStr := strings.Join(leftVals, ",")
		rightStr := strings.Join(rightVals, ",")

		switch {
		case len(leftVals) > 0 && len(rightVals) == 0:
			entry.Left = leftStr
			entry.Status = StatusRemoved
		case len(leftVals) == 0 && len(rightVals) > 0:
			entry.Right = rightStr
			entry.Status = StatusAdded
		case leftStr != rightStr:
			entry.Left = leftStr
			entry.Right = rightStr
			entry.Status = StatusModified
		default:
			entry.Left = leftStr
			entry.Status = StatusUnchanged
		}

		result.Entries = append(result.Entries, entry)
	}

	// 排序
	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Path < result.Entries[j].Path
	})

	return result
}
