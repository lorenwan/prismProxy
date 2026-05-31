package script

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ScriptStore 脚本存储
type ScriptStore struct {
	db *sql.DB
}

// NewScriptStore 创建新的脚本存储
func NewScriptStore(db *sql.DB) *ScriptStore {
	return &ScriptStore{db: db}
}

// Init 初始化脚本表
func (s *ScriptStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS scripts (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255),
		content TEXT,
		phase VARCHAR(20),
		enabled BOOLEAN DEFAULT TRUE,
		priority INTEGER DEFAULT 0,
		language VARCHAR(20) DEFAULT 'expr',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_scripts_enabled ON scripts(enabled);
	CREATE INDEX IF NOT EXISTS idx_scripts_phase ON scripts(phase);
	`
	_, err := s.db.Exec(query)
	return err
}

// List 获取脚本列表
func (s *ScriptStore) List() ([]*Script, error) {
	query := `
		SELECT id, name, content, phase, enabled, priority, language, created_at, updated_at
		FROM scripts
		ORDER BY priority ASC, created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询脚本列表失败: %w", err)
	}
	defer rows.Close()

	var scripts []*Script
	for rows.Next() {
		script := &Script{}
		err := rows.Scan(
			&script.ID, &script.Name, &script.Content,
			&script.Phase, &script.Enabled, &script.Priority,
			&script.Language, &script.CreatedAt, &script.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描脚本失败: %w", err)
		}
		scripts = append(scripts, script)
	}

	return scripts, nil
}

// Get 根据 ID 获取脚本
func (s *ScriptStore) Get(id string) (*Script, error) {
	query := `
		SELECT id, name, content, phase, enabled, priority, language, created_at, updated_at
		FROM scripts WHERE id = ?
	`

	script := &Script{}
	err := s.db.QueryRow(query, id).Scan(
		&script.ID, &script.Name, &script.Content,
		&script.Phase, &script.Enabled, &script.Priority,
		&script.Language, &script.CreatedAt, &script.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询脚本失败: %w", err)
	}

	return script, nil
}

// Create 创建脚本
func (s *ScriptStore) Create(script *Script) error {
	if script.ID == "" {
		script.ID = uuid.New().String()
	}

	now := time.Now()
	script.CreatedAt = now
	script.UpdatedAt = now

	query := `
		INSERT INTO scripts (id, name, content, phase, enabled, priority, language, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		script.ID, script.Name, script.Content,
		script.Phase, script.Enabled, script.Priority,
		script.Language, script.CreatedAt, script.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("创建脚本失败: %w", err)
	}

	return nil
}

// Update 更新脚本
func (s *ScriptStore) Update(script *Script) error {
	script.UpdatedAt = time.Now()

	query := `
		UPDATE scripts
		SET name = ?, content = ?, phase = ?, enabled = ?, priority = ?, language = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		script.Name, script.Content, script.Phase,
		script.Enabled, script.Priority, script.Language,
		script.UpdatedAt, script.ID,
	)
	if err != nil {
		return fmt.Errorf("更新脚本失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("脚本不存在")
	}

	return nil
}

// Delete 删除脚本
func (s *ScriptStore) Delete(id string) error {
	query := "DELETE FROM scripts WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除脚本失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("脚本不存在")
	}

	return nil
}

// Toggle 启用/禁用脚本
func (s *ScriptStore) Toggle(id string) (bool, error) {
	script, err := s.Get(id)
	if err != nil {
		return false, err
	}
	if script == nil {
		return false, fmt.Errorf("脚本不存在")
	}

	script.Enabled = !script.Enabled
	script.UpdatedAt = time.Now()

	query := "UPDATE scripts SET enabled = ?, updated_at = ? WHERE id = ?"
	_, err = s.db.Exec(query, script.Enabled, script.UpdatedAt, id)
	if err != nil {
		return false, fmt.Errorf("切换脚本状态失败: %w", err)
	}

	return script.Enabled, nil
}

// GetByPhase 获取指定阶段的脚本
func (s *ScriptStore) GetByPhase(phase ScriptPhase) ([]*Script, error) {
	query := `
		SELECT id, name, content, phase, enabled, priority, language, created_at, updated_at
		FROM scripts
		WHERE phase = ? AND enabled = TRUE
		ORDER BY priority ASC
	`

	rows, err := s.db.Query(query, phase)
	if err != nil {
		return nil, fmt.Errorf("查询脚本失败: %w", err)
	}
	defer rows.Close()

	var scripts []*Script
	for rows.Next() {
		script := &Script{}
		err := rows.Scan(
			&script.ID, &script.Name, &script.Content,
			&script.Phase, &script.Enabled, &script.Priority,
			&script.Language, &script.CreatedAt, &script.UpdatedAt,
		)
		if err != nil {
			continue
		}
		scripts = append(scripts, script)
	}

	return scripts, nil
}
