package search

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FilterStore 过滤器存储
type FilterStore struct {
	db *sql.DB
}

// NewFilterStore 创建过滤器存储
func NewFilterStore(db *sql.DB) *FilterStore {
	return &FilterStore{db: db}
}

// Init 初始化过滤器表
func (s *FilterStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS saved_filters (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255),
		query TEXT,
		filters TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(query)
	return err
}

// ListSavedFilters 列出所有保存的过滤器
func (s *FilterStore) ListSavedFilters() ([]SavedFilter, error) {
	query := `
		SELECT id, name, query, filters, created_at
		FROM saved_filters
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询过滤器失败: %w", err)
	}
	defer rows.Close()

	var filters []SavedFilter
	for rows.Next() {
		f := SavedFilter{}
		var filtersJSON string

		err := rows.Scan(&f.ID, &f.Name, &f.Query, &filtersJSON, &f.CreatedAt)
		if err != nil {
			continue
		}

		// 解析 filters JSON
		if filtersJSON != "" {
			json.Unmarshal([]byte(filtersJSON), &f.Filters)
		}

		filters = append(filters, f)
	}

	return filters, nil
}

// GetSavedFilter 获取指定过滤器
func (s *FilterStore) GetSavedFilter(id string) (*SavedFilter, error) {
	query := `
		SELECT id, name, query, filters, created_at
		FROM saved_filters WHERE id = ?
	`

	f := SavedFilter{}
	var filtersJSON string

	err := s.db.QueryRow(query, id).Scan(&f.ID, &f.Name, &f.Query, &filtersJSON, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询过滤器失败: %w", err)
	}

	// 解析 filters JSON
	if filtersJSON != "" {
		json.Unmarshal([]byte(filtersJSON), &f.Filters)
	}

	return &f, nil
}

// SaveFilter 保存过滤器
func (s *FilterStore) SaveFilter(f SavedFilter) (*SavedFilter, error) {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}

	f.CreatedAt = time.Now()

	// 序列化 filters
	filtersJSON, err := json.Marshal(f.Filters)
	if err != nil {
		return nil, fmt.Errorf("序列化过滤器失败: %w", err)
	}

	query := `
		INSERT INTO saved_filters (id, name, query, filters, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, f.ID, f.Name, f.Query, string(filtersJSON), f.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("保存过滤器失败: %w", err)
	}

	return &f, nil
}

// UpdateFilter 更新过滤器
func (s *FilterStore) UpdateFilter(f SavedFilter) error {
	// 序列化 filters
	filtersJSON, err := json.Marshal(f.Filters)
	if err != nil {
		return fmt.Errorf("序列化过滤器失败: %w", err)
	}

	query := `
		UPDATE saved_filters
		SET name = ?, query = ?, filters = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query, f.Name, f.Query, string(filtersJSON), f.ID)
	if err != nil {
		return fmt.Errorf("更新过滤器失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("过滤器不存在")
	}

	return nil
}

// DeleteFilter 删除过滤器
func (s *FilterStore) DeleteFilter(id string) error {
	query := "DELETE FROM saved_filters WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除过滤器失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("过滤器不存在")
	}

	return nil
}

// Count 获取过滤器数量
func (s *FilterStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM saved_filters").Scan(&count)
	return count, err
}

// SearchByName 按名称搜索过滤器
func (s *FilterStore) SearchByName(name string) ([]SavedFilter, error) {
	query := `
		SELECT id, name, query, filters, created_at
		FROM saved_filters
		WHERE name LIKE ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("搜索过滤器失败: %w", err)
	}
	defer rows.Close()

	var filters []SavedFilter
	for rows.Next() {
		f := SavedFilter{}
		var filtersJSON string

		err := rows.Scan(&f.ID, &f.Name, &f.Query, &filtersJSON, &f.CreatedAt)
		if err != nil {
			continue
		}

		if filtersJSON != "" {
			json.Unmarshal([]byte(filtersJSON), &f.Filters)
		}

		filters = append(filters, f)
	}

	return filters, nil
}
