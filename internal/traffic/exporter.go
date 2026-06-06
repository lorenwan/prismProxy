package traffic

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Exporter 流量导出器
type Exporter struct {
	manager *Manager
}

// NewExporter 创建新的导出器
func NewExporter(manager *Manager) *Exporter {
	return &Exporter{manager: manager}
}

// ExportFormat 导出格式
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatHAR  ExportFormat = "har"
)

// ExportJSON 导出为 JSON
func (e *Exporter) ExportJSON(transactions []*Transaction) ([]byte, error) {
	return json.MarshalIndent(transactions, "", "  ")
}

// ExportCSV 导出为 CSV
func (e *Exporter) ExportCSV(transactions []*Transaction) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// 写入表头
	writer.Write([]string{
		"ID", "时间", "方法", "URL", "主机", "状态码", "耗时(ms)",
		"请求大小", "响应大小", "内容类型",
	})

	// 写入数据
	for _, tx := range transactions {
		writer.Write([]string{
			fmt.Sprintf("%d", tx.ID),
			tx.Timestamp.Format(time.RFC3339),
			tx.Method,
			tx.URL,
			tx.Host,
			fmt.Sprintf("%d", tx.Response.StatusCode),
			fmt.Sprintf("%d", tx.DurationMs),
			fmt.Sprintf("%d", tx.Request.BodySize),
			fmt.Sprintf("%d", tx.Response.BodySize),
			tx.Request.ContentType,
		})
	}

	writer.Flush()
	return []byte(buf.String()), writer.Error()
}

// ExportHAR 导出为 HAR 格式
func (e *Exporter) ExportHAR(transactions []*Transaction) ([]byte, error) {
	har := HAR{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "PrismProxy",
				Version: "1.0.0",
			},
			Entries: make([]HAREntry, len(transactions)),
		},
	}

	for i, tx := range transactions {
		entry := HAREntry{
			StartedDateTime: tx.Timestamp.Format(time.RFC3339),
			Time:            float64(tx.DurationMs),
			Request: HARRequest{
				Method:      tx.Method,
				URL:         tx.URL,
				HTTPVersion: "HTTP/1.1",
				Headers:     convertHeaders(tx.Request.Headers),
				HeadersSize: -1,
				BodySize:    tx.Request.BodySize,
			},
			Response: HARResponse{
				Status:      tx.Response.StatusCode,
				StatusText:  tx.Response.StatusText,
				HTTPVersion: "HTTP/1.1",
				Headers:     convertHeaders(tx.Response.Headers),
				HeadersSize: -1,
				BodySize:    tx.Response.BodySize,
				Content: HARContent{
					Size:     tx.Response.BodySize,
					MimeType: tx.Response.ContentType,
				},
			},
		}
		har.Log.Entries[i] = entry
	}

	return json.MarshalIndent(har, "", "  ")
}

// HAR 数据结构
type HAR struct {
	Log HARLog `json:"log"`
}

type HARLog struct {
	Version string     `json:"version"`
	Creator HARCreator `json:"creator"`
	Entries []HAREntry `json:"entries"`
}

type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type HAREntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
}

type HARRequest struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []HARHeader `json:"headers"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

type HARResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []HARHeader `json:"headers"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
	Content     HARContent  `json:"content"`
}

type HARContent struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// convertHeaders 转换 headers 格式
func convertHeaders(headers map[string][]string) []HARHeader {
	var result []HARHeader
	for name, values := range headers {
		for _, value := range values {
			result = append(result, HARHeader{
				Name:  name,
				Value: value,
			})
		}
	}
	return result
}
