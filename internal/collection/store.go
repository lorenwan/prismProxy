package collection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store 集合存储
type Store struct {
	db *sql.DB
}

// NewStore 创建新的集合存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init 初始化集合表
func (s *Store) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS collections (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		parent_id VARCHAR(36),
		config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (parent_id) REFERENCES collections(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_collections_parent ON collections(parent_id);

	CREATE TABLE IF NOT EXISTS collection_items (
		id VARCHAR(36) PRIMARY KEY,
		collection_id VARCHAR(36) NOT NULL,
		parent_id VARCHAR(36),
		type VARCHAR(20) NOT NULL DEFAULT 'request',
		name VARCHAR(255) NOT NULL,
		request_config TEXT,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES collection_items(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_items_collection ON collection_items(collection_id);
	CREATE INDEX IF NOT EXISTS idx_items_parent ON collection_items(parent_id);
	CREATE INDEX IF NOT EXISTS idx_items_type ON collection_items(type);
	`
	_, err := s.db.Exec(query)
	return err
}

// Create 创建集合
func (s *Store) Create(col *Collection) error {
	if col.ID == "" {
		col.ID = uuid.New().String()
	}
	now := time.Now()
	col.CreatedAt = now
	col.UpdatedAt = now

	query := `
		INSERT INTO collections (id, name, description, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		col.ID, col.Name, col.Description,
		nullString(col.ParentID),
		col.CreatedAt, col.UpdatedAt,
	)
	return err
}

// GetByID 根据 ID 获取集合
func (s *Store) GetByID(id string) (*Collection, error) {
	query := `
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM collections WHERE id = ?
	`
	col := &Collection{}
	var parentID sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&col.ID, &col.Name, &col.Description,
		&parentID, &col.CreatedAt, &col.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询集合失败: %w", err)
	}
	col.ParentID = parentID.String

	// 获取子项目
	items, err := s.GetItems(id, "")
	if err != nil {
		return nil, err
	}
	col.Items = items

	return col, nil
}

// List 获取集合列表
func (s *Store) List() ([]*Collection, error) {
	query := `
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM collections
		WHERE parent_id IS NULL
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询集合列表失败: %w", err)
	}
	defer rows.Close()

	var collections []*Collection
	for rows.Next() {
		col := &Collection{}
		var parentID sql.NullString

		err := rows.Scan(
			&col.ID, &col.Name, &col.Description,
			&parentID, &col.CreatedAt, &col.UpdatedAt,
		)
		if err != nil {
			continue
		}
		col.ParentID = parentID.String
		collections = append(collections, col)
	}

	return collections, nil
}

// Update 更新集合
func (s *Store) Update(col *Collection) error {
	col.UpdatedAt = time.Now()

	query := `
		UPDATE collections
		SET name = ?, description = ?, parent_id = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query,
		col.Name, col.Description,
		nullString(col.ParentID),
		col.UpdatedAt, col.ID,
	)
	if err != nil {
		return fmt.Errorf("更新集合失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("集合不存在")
	}

	return nil
}

// Delete 删除集合
func (s *Store) Delete(id string) error {
	query := "DELETE FROM collections WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除集合失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("集合不存在")
	}

	return nil
}

// GetItems 获取集合项目列表
func (s *Store) GetItems(collectionID, parentID string) ([]CollectionItem, error) {
	var query string
	var args []interface{}

	if parentID == "" {
		query = `
			SELECT id, type, name, request_config, created_at, updated_at
			FROM collection_items
			WHERE collection_id = ? AND parent_id IS NULL
			ORDER BY sort_order ASC, created_at ASC
		`
		args = []interface{}{collectionID}
	} else {
		query = `
			SELECT id, type, name, request_config, created_at, updated_at
			FROM collection_items
			WHERE collection_id = ? AND parent_id = ?
			ORDER BY sort_order ASC, created_at ASC
		`
		args = []interface{}{collectionID, parentID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询项目列表失败: %w", err)
	}
	defer rows.Close()

	var items []CollectionItem
	for rows.Next() {
		item := CollectionItem{}
		var reqJSON sql.NullString

		err := rows.Scan(
			&item.ID, &item.Type, &item.Name,
			&reqJSON, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if reqJSON.Valid && item.Type == "request" {
			var req APIRequest
			if err := json.Unmarshal([]byte(reqJSON.String), &req); err == nil {
				item.Request = &req
			}
		}

		// 如果是文件夹，递归获取子项目
		if item.Type == "folder" {
			subItems, err := s.GetItems(collectionID, item.ID)
			if err == nil {
				item.Items = subItems
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// CreateItem 创建集合项目
func (s *Store) CreateItem(collectionID, parentID string, item *CollectionItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	var reqJSON string
	if item.Request != nil {
		data, err := json.Marshal(item.Request)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		reqJSON = string(data)
	}

	query := `
		INSERT INTO collection_items (id, collection_id, parent_id, type, name, request_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		item.ID, collectionID,
		nullString(parentID),
		item.Type, item.Name, reqJSON,
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

// UpdateItem 更新集合项目
func (s *Store) UpdateItem(item *CollectionItem) error {
	item.UpdatedAt = time.Now()

	var reqJSON string
	if item.Request != nil {
		data, err := json.Marshal(item.Request)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		reqJSON = string(data)
	}

	query := `
		UPDATE collection_items
		SET name = ?, request_config = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query, item.Name, reqJSON, item.UpdatedAt, item.ID)
	if err != nil {
		return fmt.Errorf("更新项目失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("项目不存在")
	}

	return nil
}

// DeleteItem 删除集合项目
func (s *Store) DeleteItem(id string) error {
	query := "DELETE FROM collection_items WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("项目不存在")
	}

	return nil
}

// GetItemByID 根据 ID 获取项目
func (s *Store) GetItemByID(id string) (*CollectionItem, error) {
	query := `
		SELECT id, collection_id, type, name, request_config, created_at, updated_at
		FROM collection_items WHERE id = ?
	`
	item := &CollectionItem{}
	var collectionID string
	var reqJSON sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&item.ID, &collectionID, &item.Type, &item.Name,
		&reqJSON, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	if reqJSON.Valid && item.Type == "request" {
		var req APIRequest
		if err := json.Unmarshal([]byte(reqJSON.String), &req); err == nil {
			item.Request = &req
		}
	}

	return item, nil
}

// nullString 将空字符串转换为 NULL
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
