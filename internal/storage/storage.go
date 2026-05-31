package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// TrafficData 抓包数据结构
type TrafficData struct {
	ID           int64  `json:"id"`
	Method       string `json:"method"`
	URL          string `json:"url"`
	Host         string `json:"host"`
	ServerIP     string `json:"server_ip"`
	ContentType  string `json:"content_type"`
	StatusCode   int    `json:"status_code"`
	DurationMs   int64  `json:"duration_ms"`
	ReqHeaders   string `json:"req_headers"`
	ReqBody      []byte `json:"req_body"`
	RespHeaders  string `json:"resp_headers"`
	RespBody     []byte `json:"resp_body"`
	RawReq       []byte `json:"raw_req"`
	RawResp      []byte `json:"raw_resp"`
	Timestamp    string `json:"timestamp"`
}

// Storage SQLite 存储层
type Storage struct {
	DB *sql.DB
}

// NewStorage 创建新的存储实例
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 启用 WAL 模式提高并发性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置 WAL 模式失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	log.Println("[INFO] SQLite 数据库连接成功")
	return &Storage{DB: db}, nil
}

// Close 关闭数据库连接
func (s *Storage) Close() {
	if s.DB != nil {
		s.DB.Close()
		log.Println("[INFO] 数据库连接已关闭")
	}
}

// SaveTraffic 保存抓包数据
func (s *Storage) SaveTraffic(data *TrafficData) error {
	query := `
	INSERT INTO traffic (method, url, host, server_ip, content_type, status_code,
		duration_ms, req_headers, req_body, resp_headers, resp_body, raw_req, raw_resp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.DB.Exec(query,
		data.Method, data.URL, data.Host, data.ServerIP, data.ContentType,
		data.StatusCode, data.DurationMs, data.ReqHeaders, data.ReqBody,
		data.RespHeaders, data.RespBody, data.RawReq, data.RawResp)
	if err != nil {
		return fmt.Errorf("保存抓包数据失败: %w", err)
	}

	id, _ := result.LastInsertId()
	data.ID = id
	return nil
}

// GetTrafficByID 根据 ID 获取抓包数据
func (s *Storage) GetTrafficByID(id int64) (*TrafficData, error) {
	query := `SELECT id, method, url, host, server_ip, content_type, status_code,
		duration_ms, req_headers, req_body, resp_headers, resp_body, raw_req, raw_resp, timestamp
		FROM traffic WHERE id = ?`

	data := &TrafficData{}
	err := s.DB.QueryRow(query, id).Scan(
		&data.ID, &data.Method, &data.URL, &data.Host, &data.ServerIP,
		&data.ContentType, &data.StatusCode, &data.DurationMs,
		&data.ReqHeaders, &data.ReqBody, &data.RespHeaders, &data.RespBody,
		&data.RawReq, &data.RawResp, &data.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询抓包数据失败: %w", err)
	}
	return data, nil
}

// GetTrafficList 获取抓包数据列表
func (s *Storage) GetTrafficList(limit, offset int) ([]*TrafficData, error) {
	query := `SELECT id, method, url, host, server_ip, content_type, status_code,
		duration_ms, timestamp
		FROM traffic ORDER BY id DESC LIMIT ? OFFSET ?`

	rows, err := s.DB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询抓包列表失败: %w", err)
	}
	defer rows.Close()

	var list []*TrafficData
	for rows.Next() {
		data := &TrafficData{}
		err := rows.Scan(&data.ID, &data.Method, &data.URL, &data.Host,
			&data.ServerIP, &data.ContentType, &data.StatusCode,
			&data.DurationMs, &data.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}
		list = append(list, data)
	}
	return list, nil
}

// GetTrafficCount 获取抓包数据总数
func (s *Storage) GetTrafficCount() (int64, error) {
	var count int64
	err := s.DB.QueryRow("SELECT COUNT(*) FROM traffic").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询总数失败: %w", err)
	}
	return count, nil
}

// DeleteTrafficByID 根据 ID 删除抓包数据
func (s *Storage) DeleteTrafficByID(id int64) error {
	_, err := s.DB.Exec("DELETE FROM traffic WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除抓包数据失败: %w", err)
	}
	return nil
}

// ClearTraffic 清空所有抓包数据
func (s *Storage) ClearTraffic() error {
	_, err := s.DB.Exec("DELETE FROM traffic")
	if err != nil {
		return fmt.Errorf("清空抓包数据失败: %w", err)
	}
	return nil
}

// Settings 应用设置
type Settings struct {
	ProxyPort        int    `json:"proxy_port"`
	ProxyListenAddr  string `json:"proxy_listen_addr"`
	ProxyEnableHTTPS bool   `json:"proxy_enable_https"`
	ProxyEnableMITM  bool   `json:"proxy_enable_mitm"`
	Theme            string `json:"theme"`
	Language         string `json:"language"`
	EnableTrafficLog bool   `json:"enable_traffic_log"`
	MaxTrafficCount  int    `json:"max_traffic_count"`
	TrafficTTLHours  int    `json:"traffic_ttl_hours"`
}

// GetSettings 获取设置
func (s *Storage) GetSettings() (*Settings, error) {
	// 默认设置
	settings := &Settings{
		ProxyPort:        8888,
		ProxyListenAddr:  "0.0.0.0",
		ProxyEnableHTTPS: true,
		ProxyEnableMITM:  true,
		Theme:            "dark",
		Language:         "zh",
		EnableTrafficLog: true,
		MaxTrafficCount:  10000,
		TrafficTTLHours:  72,
	}

	rows, err := s.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		// 表可能不存在，返回默认值
		return settings, nil
	}
	defer rows.Close()

	kv := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		kv[key] = value
	}

	// 从数据库覆盖默认值
	if v, ok := kv["proxy_port"]; ok {
		fmt.Sscanf(v, "%d", &settings.ProxyPort)
	}
	if v, ok := kv["proxy_listen_addr"]; ok {
		settings.ProxyListenAddr = v
	}
	if v, ok := kv["proxy_enable_https"]; ok {
		settings.ProxyEnableHTTPS = v == "true"
	}
	if v, ok := kv["proxy_enable_mitm"]; ok {
		settings.ProxyEnableMITM = v == "true"
	}
	if v, ok := kv["theme"]; ok {
		settings.Theme = v
	}
	if v, ok := kv["language"]; ok {
		settings.Language = v
	}
	if v, ok := kv["enable_traffic_log"]; ok {
		settings.EnableTrafficLog = v == "true"
	}
	if v, ok := kv["max_traffic_count"]; ok {
		fmt.Sscanf(v, "%d", &settings.MaxTrafficCount)
	}
	if v, ok := kv["traffic_ttl_hours"]; ok {
		fmt.Sscanf(v, "%d", &settings.TrafficTTLHours)
	}

	return settings, nil
}

// SaveSettings 保存设置
func (s *Storage) SaveSettings(settings *Settings) error {
	// 将设置逐个写入 key-value 表
	pairs := map[string]string{
		"proxy_port":         fmt.Sprintf("%d", settings.ProxyPort),
		"proxy_listen_addr":  settings.ProxyListenAddr,
		"proxy_enable_https": fmt.Sprintf("%t", settings.ProxyEnableHTTPS),
		"proxy_enable_mitm":  fmt.Sprintf("%t", settings.ProxyEnableMITM),
		"theme":              settings.Theme,
		"language":           settings.Language,
		"enable_traffic_log": fmt.Sprintf("%t", settings.EnableTrafficLog),
		"max_traffic_count":  fmt.Sprintf("%d", settings.MaxTrafficCount),
		"traffic_ttl_hours":  fmt.Sprintf("%d", settings.TrafficTTLHours),
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	for key, value := range pairs {
		_, err := tx.Exec(
			"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
			key, value,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("保存设置 %s 失败: %w", key, err)
		}
	}

	return tx.Commit()
}
