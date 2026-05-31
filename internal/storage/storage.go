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
