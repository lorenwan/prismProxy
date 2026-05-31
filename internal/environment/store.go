package environment

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store 环境存储
type Store struct {
	db *sql.DB
}

// NewStore 创建新的环境存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init 初始化环境表
func (s *Store) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS environments (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		is_default BOOLEAN DEFAULT FALSE,
		base_url VARCHAR(500),
		variables_config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_environments_active ON environments(is_active);
	`
	_, err := s.db.Exec(query)
	return err
}

// Create 创建环境
func (s *Store) Create(env *Environment) error {
	if env.ID == "" {
		env.ID = uuid.New().String()
	}
	now := time.Now()
	env.CreatedAt = now
	env.UpdatedAt = now

	varsJSON, err := json.Marshal(env.Variables)
	if err != nil {
		return fmt.Errorf("序列化变量失败: %w", err)
	}

	query := `
		INSERT INTO environments (id, name, is_active, is_default, base_url, variables_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.Exec(query,
		env.ID, env.Name, env.IsActive, env.IsDefault,
		env.BaseURL, string(varsJSON),
		env.CreatedAt, env.UpdatedAt,
	)
	return err
}

// GetByID 根据 ID 获取环境
func (s *Store) GetByID(id string) (*Environment, error) {
	query := `
		SELECT id, name, is_active, is_default, base_url, variables_config, created_at, updated_at
		FROM environments WHERE id = ?
	`
	env := &Environment{}
	var varsJSON sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&env.ID, &env.Name, &env.IsActive, &env.IsDefault,
		&env.BaseURL, &varsJSON,
		&env.CreatedAt, &env.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询环境失败: %w", err)
	}

	if varsJSON.Valid {
		if err := json.Unmarshal([]byte(varsJSON.String), &env.Variables); err != nil {
			return nil, fmt.Errorf("解析变量失败: %w", err)
		}
	}

	return env, nil
}

// List 获取环境列表
func (s *Store) List() ([]*Environment, error) {
	query := `
		SELECT id, name, is_active, is_default, base_url, variables_config, created_at, updated_at
		FROM environments
		ORDER BY is_default DESC, name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询环境列表失败: %w", err)
	}
	defer rows.Close()

	var envs []*Environment
	for rows.Next() {
		env := &Environment{}
		var varsJSON sql.NullString

		err := rows.Scan(
			&env.ID, &env.Name, &env.IsActive, &env.IsDefault,
			&env.BaseURL, &varsJSON,
			&env.CreatedAt, &env.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if varsJSON.Valid {
			json.Unmarshal([]byte(varsJSON.String), &env.Variables)
		}

		envs = append(envs, env)
	}

	return envs, nil
}

// Update 更新环境
func (s *Store) Update(env *Environment) error {
	env.UpdatedAt = time.Now()

	varsJSON, err := json.Marshal(env.Variables)
	if err != nil {
		return fmt.Errorf("序列化变量失败: %w", err)
	}

	query := `
		UPDATE environments
		SET name = ?, is_active = ?, is_default = ?, base_url = ?, variables_config = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query,
		env.Name, env.IsActive, env.IsDefault,
		env.BaseURL, string(varsJSON),
		env.UpdatedAt, env.ID,
	)
	if err != nil {
		return fmt.Errorf("更新环境失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("环境不存在")
	}

	return nil
}

// Delete 删除环境
func (s *Store) Delete(id string) error {
	query := "DELETE FROM environments WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除环境失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("环境不存在")
	}

	return nil
}

// SetActive 设置活跃环境
func (s *Store) SetActive(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// 取消所有环境的活跃状态
	_, err = tx.Exec("UPDATE environments SET is_active = FALSE")
	if err != nil {
		tx.Rollback()
		return err
	}

	// 设置指定环境为活跃
	result, err := tx.Exec("UPDATE environments SET is_active = TRUE WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		tx.Rollback()
		return fmt.Errorf("环境不存在")
	}

	return tx.Commit()
}

// GetActive 获取活跃环境
func (s *Store) GetActive() (*Environment, error) {
	query := `
		SELECT id, name, is_active, is_default, base_url, variables_config, created_at, updated_at
		FROM environments WHERE is_active = TRUE LIMIT 1
	`
	env := &Environment{}
	var varsJSON sql.NullString

	err := s.db.QueryRow(query).Scan(
		&env.ID, &env.Name, &env.IsActive, &env.IsDefault,
		&env.BaseURL, &varsJSON,
		&env.CreatedAt, &env.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询活跃环境失败: %w", err)
	}

	if varsJSON.Valid {
		json.Unmarshal([]byte(varsJSON.String), &env.Variables)
	}

	return env, nil
}

// GetVariables 获取环境变量映射
func (s *Store) GetVariables(id string) (map[string]string, error) {
	env, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("环境不存在")
	}

	vars := make(map[string]string)
	for _, v := range env.Variables {
		if v.Enabled {
			vars[v.Key] = v.Value
		}
	}

	// 添加内置变量
	vars["environment.name"] = env.Name
	if env.BaseURL != "" {
		vars["base_url"] = env.BaseURL
	}

	return vars, nil
}
