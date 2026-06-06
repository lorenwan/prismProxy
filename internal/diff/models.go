package diff

// DiffResult 对比结果
type DiffResult struct {
	Type    DiffType    `json:"type"`
	Entries []DiffEntry `json:"entries"`
}

// DiffType 对比类型
type DiffType string

const (
	DiffTypeHeaders DiffType = "headers"
	DiffTypeBody    DiffType = "body"
	DiffTypeJSON    DiffType = "json"
	DiffTypeQuery   DiffType = "query"
)

// DiffEntry 对比条目
type DiffEntry struct {
	Path   string     `json:"path"`
	Left   string     `json:"left,omitempty"`
	Right  string     `json:"right,omitempty"`
	Status DiffStatus `json:"status"`
}

// DiffStatus 对比状态
type DiffStatus string

const (
	StatusAdded     DiffStatus = "added"
	StatusRemoved   DiffStatus = "removed"
	StatusModified  DiffStatus = "modified"
	StatusUnchanged DiffStatus = "unchanged"
)

// JSONDiffResult JSON 对比结果
type JSONDiffResult struct {
	Diffs   []JSONDiffEntry `json:"diffs"`
	Summary DiffSummary     `json:"summary"`
}

// JSONDiffEntry JSON 对比条目
type JSONDiffEntry struct {
	Path   string      `json:"path"`
	Left   interface{} `json:"left,omitempty"`
	Right  interface{} `json:"right,omitempty"`
	Status DiffStatus  `json:"status"`
	Type   string      `json:"type"` // string, number, bool, array, object, null
}

// DiffSummary 对比摘要
type DiffSummary struct {
	TotalFields int `json:"total_fields"`
	Added       int `json:"added"`
	Removed     int `json:"removed"`
	Modified    int `json:"modified"`
	Unchanged   int `json:"unchanged"`
}
