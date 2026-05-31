package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CompareFlatMap 对比扁平 Map
func (e *DiffEngine) CompareFlatMap(left, right map[string]string) DiffResult {
	result := DiffResult{
		Type:    DiffTypeHeaders,
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
		leftVal, leftOk := left[key]
		rightVal, rightOk := right[key]

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

// CompareStrings 对比字符串
func (e *DiffEngine) CompareStrings(left, right string) DiffResult {
	result := DiffResult{
		Type:    DiffTypeBody,
		Entries: make([]DiffEntry, 0),
	}

	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	// 使用简单的逐行对比
	maxLen := len(leftLines)
	if len(rightLines) > maxLen {
		maxLen = len(rightLines)
	}

	for i := 0; i < maxLen; i++ {
		entry := DiffEntry{
			Path: fmt.Sprintf("line:%d", i+1),
		}

		switch {
		case i >= len(leftLines):
			entry.Right = rightLines[i]
			entry.Status = StatusAdded
		case i >= len(rightLines):
			entry.Left = leftLines[i]
			entry.Status = StatusRemoved
		case leftLines[i] != rightLines[i]:
			entry.Left = leftLines[i]
			entry.Right = rightLines[i]
			entry.Status = StatusModified
		default:
			entry.Left = leftLines[i]
			entry.Status = StatusUnchanged
		}

		result.Entries = append(result.Entries, entry)
	}

	return result
}

// CompareJSONStrings 对比 JSON 字符串
func (e *DiffEngine) CompareJSONStrings(left, right string) (JSONDiffResult, error) {
	var leftObj, rightObj interface{}

	if err := json.Unmarshal([]byte(left), &leftObj); err != nil {
		return JSONDiffResult{}, fmt.Errorf("解析左侧 JSON 失败: %w", err)
	}

	if err := json.Unmarshal([]byte(right), &rightObj); err != nil {
		return JSONDiffResult{}, fmt.Errorf("解析右侧 JSON 失败: %w", err)
	}

	return e.CompareJSON([]byte(left), []byte(right)), nil
}

// GetDiffStats 获取差异统计
func (e *DiffEngine) GetDiffStats(result DiffResult) map[string]int {
	stats := map[string]int{
		"total":     len(result.Entries),
		"added":     0,
		"removed":   0,
		"modified":  0,
		"unchanged": 0,
	}

	for _, entry := range result.Entries {
		switch entry.Status {
		case StatusAdded:
			stats["added"]++
		case StatusRemoved:
			stats["removed"]++
		case StatusModified:
			stats["modified"]++
		case StatusUnchanged:
			stats["unchanged"]++
		}
	}

	return stats
}

// FilterDiffs 过滤差异（只返回非 unchanged）
func (e *DiffEngine) FilterDiffs(result DiffResult) DiffResult {
	filtered := DiffResult{
		Type:    result.Type,
		Entries: make([]DiffEntry, 0),
	}

	for _, entry := range result.Entries {
		if entry.Status != StatusUnchanged {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}

	return filtered
}

// HasDiffs 是否有差异
func (e *DiffEngine) HasDiffs(result DiffResult) bool {
	for _, entry := range result.Entries {
		if entry.Status != StatusUnchanged {
			return true
		}
	}
	return false
}

// GetJSONDiffStats 获取 JSON 差异统计
func (e *DiffEngine) GetJSONDiffStats(result JSONDiffResult) map[string]int {
	return map[string]int{
		"total_fields": result.Summary.TotalFields,
		"added":        result.Summary.Added,
		"removed":      result.Summary.Removed,
		"modified":     result.Summary.Modified,
		"unchanged":    result.Summary.Unchanged,
	}
}
