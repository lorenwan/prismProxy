package traffic

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"prismproxy/internal/storage"
	"prismproxy/internal/websocket"
)

// Manager 流量管理器
type Manager struct {
	storage    *storage.Storage
	hub        *websocket.Hub
	mu         sync.RWMutex
	collectors []*Collector
}

// NewManager 创建新的流量管理器
func NewManager(store *storage.Storage, hub *websocket.Hub) *Manager {
	return &Manager{
		storage: store,
		hub:     hub,
	}
}

// SaveTransaction 保存流量记录
func (m *Manager) SaveTransaction(tx *Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 转换为存储格式
	data := &storage.TrafficData{
		Method:      tx.Method,
		URL:         tx.URL,
		Host:        tx.Host,
		ServerIP:    tx.ServerIP,
		ContentType: tx.Request.ContentType,
		StatusCode:  tx.Response.StatusCode,
		DurationMs:  tx.DurationMs,
		ReqHeaders:  marshalHeaders(tx.Request.Headers),
		ReqBody:     tx.Request.Body,
		RespHeaders: marshalHeaders(tx.Response.Headers),
		RespBody:    tx.Response.Body,
		RawReq:      tx.Request.Raw,
		RawResp:     tx.Response.Raw,
	}

	if err := m.storage.SaveTraffic(data); err != nil {
		return fmt.Errorf("保存流量记录失败: %w", err)
	}

	tx.ID = data.ID

	// 通过 WebSocket 广播新流量
	if m.hub != nil {
		m.hub.Broadcast(&websocket.Message{
			Type:    "traffic:new",
			Payload: tx,
		})
	}

	return nil
}

// GetTransaction 获取单条流量记录
func (m *Manager) GetTransaction(id int64) (*Transaction, error) {
	data, err := m.storage.GetTrafficByID(id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	return m.convertToTransaction(data), nil
}

// ListTransactions 获取流量列表
func (m *Manager) ListTransactions(limit, offset int) ([]*Transaction, int64, error) {
	list, err := m.storage.GetTrafficList(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.storage.GetTrafficCount()
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*Transaction, len(list))
	for i, data := range list {
		transactions[i] = m.convertToTransaction(data)
	}

	return transactions, count, nil
}

// ListWithFilter 带过滤器的列表查询
func (m *Manager) ListWithFilter(filter *Filter, limit, offset int) ([]*Transaction, int64, error) {
	where, args := filter.BuildSQL()

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traffic %s", where)
	var count int64
	err := m.storage.DB.QueryRow(countQuery, args...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	listQuery := fmt.Sprintf(`
		SELECT id, method, url, host, server_ip, content_type, status_code,
			duration_ms, timestamp
		FROM traffic %s
		ORDER BY id DESC LIMIT ? OFFSET ?
	`, where)

	args = append(args, limit, offset)
	rows, err := m.storage.DB.Query(listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(&tx.ID, &tx.Method, &tx.URL, &tx.Host,
			&tx.ServerIP, &tx.Request.ContentType, &tx.Response.StatusCode,
			&tx.DurationMs, &tx.Timestamp)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描数据失败: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, count, nil
}

// DeleteTransaction 删除流量记录
func (m *Manager) DeleteTransaction(id int64) error {
	if err := m.storage.DeleteTrafficByID(id); err != nil {
		return err
	}

	// 广播删除事件
	if m.hub != nil {
		m.hub.Broadcast(&websocket.Message{
			Type:    "traffic:delete",
			Payload: map[string]int64{"id": id},
		})
	}

	return nil
}

// ClearTransactions 清空流量记录
func (m *Manager) ClearTransactions() error {
	if err := m.storage.ClearTraffic(); err != nil {
		return err
	}

	// 广播清空事件
	if m.hub != nil {
		m.hub.Broadcast(&websocket.Message{
			Type:    "traffic:clear",
			Payload: nil,
		})
	}

	return nil
}

// UpdateBookmark 更新书签状态
func (m *Manager) UpdateBookmark(id int64, bookmarked bool) error {
	query := "UPDATE traffic SET bookmarked = ? WHERE id = ?"
	_, err := m.storage.DB.Exec(query, bookmarked, id)
	return err
}

// UpdateNotes 更新备注
func (m *Manager) UpdateNotes(id int64, notes string) error {
	query := "UPDATE traffic SET notes = ? WHERE id = ?"
	_, err := m.storage.DB.Exec(query, notes, id)
	return err
}

// UpdateColor 更新颜色标记
func (m *Manager) UpdateColor(id int64, color string) error {
	query := "UPDATE traffic SET color = ? WHERE id = ?"
	_, err := m.storage.DB.Exec(query, color, id)
	return err
}

// UpdateTags 更新标签
func (m *Manager) UpdateTags(id int64, tags []string) error {
	tagsJSON, _ := json.Marshal(tags)
	query := "UPDATE traffic SET tags = ? WHERE id = ?"
	_, err := m.storage.DB.Exec(query, string(tagsJSON), id)
	return err
}

// GetStats 获取流量统计
func (m *Manager) GetStats() (*TrafficStats, error) {
	stats := &TrafficStats{}

	// 总数和平均耗时
	err := m.storage.DB.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(duration_ms), 0),
			COALESCE(MAX(duration_ms), 0), COALESCE(MIN(duration_ms), 0)
		FROM traffic
	`).Scan(&stats.TotalRequests, &stats.AvgDuration, &stats.MaxDuration, &stats.MinDuration)
	if err != nil {
		return nil, err
	}

	// 成功/错误计数
	m.storage.DB.QueryRow("SELECT COUNT(*) FROM traffic WHERE status_code >= 200 AND status_code < 400").Scan(&stats.SuccessCount)
	m.storage.DB.QueryRow("SELECT COUNT(*) FROM traffic WHERE status_code >= 400").Scan(&stats.ErrorCount)

	// 主机统计
	rows, err := m.storage.DB.Query(`
		SELECT host, COUNT(*) as cnt, AVG(duration_ms)
		FROM traffic
		GROUP BY host
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var hs HostStat
			rows.Scan(&hs.Host, &hs.Count, &hs.AvgTime)
			stats.HostStats = append(stats.HostStats, hs)
		}
	}

	// 方法统计
	rows, err = m.storage.DB.Query(`
		SELECT method, COUNT(*) as cnt
		FROM traffic
		GROUP BY method
		ORDER BY cnt DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ms MethodStat
			rows.Scan(&ms.Method, &ms.Count)
			stats.MethodStats = append(stats.MethodStats, ms)
		}
	}

	// 状态码统计
	rows, err = m.storage.DB.Query(`
		SELECT status_code, COUNT(*) as cnt
		FROM traffic
		GROUP BY status_code
		ORDER BY cnt DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ss StatusStat
			rows.Scan(&ss.StatusCode, &ss.Count)
			stats.StatusStats = append(stats.StatusStats, ss)
		}
	}

	return stats, nil
}

// RegisterCollector 注册流量收集器
func (m *Manager) RegisterCollector(c *Collector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectors = append(m.collectors, c)
}

// convertToTransaction 转换为 Transaction
func (m *Manager) convertToTransaction(data *storage.TrafficData) *Transaction {
	return &Transaction{
		ID:         data.ID,
		Timestamp:  parseTime(data.Timestamp),
		DurationMs: data.DurationMs,
		Method:     data.Method,
		URL:        data.URL,
		Host:       data.Host,
		ServerIP:   data.ServerIP,
		Request: &RequestData{
			ContentType: data.ContentType,
			Body:        data.ReqBody,
			Raw:         data.RawReq,
		},
		Response: &ResponseData{
			StatusCode:  data.StatusCode,
			ContentType: data.ContentType,
			Body:        data.RespBody,
			Raw:         data.RawResp,
		},
	}
}

// marshalHeaders 序列化 headers
func marshalHeaders(headers map[string][]string) string {
	if headers == nil {
		return "{}"
	}
	data, _ := json.Marshal(headers)
	return string(data)
}

// parseTime 解析时间字符串
func parseTime(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Now()
	}
	return t
}
