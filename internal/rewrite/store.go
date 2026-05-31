package rewrite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store 重写规则存储
type Store struct {
	db *sql.DB
}

// NewStore 创建新的重写规则存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init 初始化重写规则表
func (s *Store) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS rewrite_rules (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255),
		enabled BOOLEAN DEFAULT TRUE,
		priority INTEGER DEFAULT 0,
		match_config TEXT,
		actions_config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_rewrite_rules_enabled ON rewrite_rules(enabled);
	CREATE INDEX IF NOT EXISTS idx_rewrite_rules_priority ON rewrite_rules(priority);
	`
	_, err := s.db.Exec(query)
	return err
}

// Create 创建重写规则
func (s *Store) Create(rule *RewriteRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	matchJSON, err := json.Marshal(rule.Match)
	if err != nil {
		return fmt.Errorf("序列化匹配条件失败: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	query := `
		INSERT INTO rewrite_rules (id, name, enabled, priority, match_config, actions_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		rule.ID, rule.Name, rule.Enabled, rule.Priority,
		string(matchJSON), string(actionsJSON),
		rule.CreatedAt, rule.UpdatedAt,
	)

	return err
}

// GetByID 根据 ID 获取重写规则
func (s *Store) GetByID(id string) (*RewriteRule, error) {
	query := `
		SELECT id, name, enabled, priority, match_config, actions_config, created_at, updated_at
		FROM rewrite_rules WHERE id = ?
	`

	rule := &RewriteRule{}
	var matchJSON, actionsJSON string

	err := s.db.QueryRow(query, id).Scan(
		&rule.ID, &rule.Name, &rule.Enabled, &rule.Priority,
		&matchJSON, &actionsJSON,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询重写规则失败: %w", err)
	}

	if err := json.Unmarshal([]byte(matchJSON), &rule.Match); err != nil {
		return nil, fmt.Errorf("解析匹配条件失败: %w", err)
	}

	if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
		return nil, fmt.Errorf("解析动作失败: %w", err)
	}

	return rule, nil
}

// List 获取重写规则列表
func (s *Store) List() ([]*RewriteRule, error) {
	query := `
		SELECT id, name, enabled, priority, match_config, actions_config, created_at, updated_at
		FROM rewrite_rules
		ORDER BY priority ASC, created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询重写规则列表失败: %w", err)
	}
	defer rows.Close()

	var rules []*RewriteRule
	for rows.Next() {
		rule := &RewriteRule{}
		var matchJSON, actionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Enabled, &rule.Priority,
			&matchJSON, &actionsJSON,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描重写规则失败: %w", err)
		}

		if err := json.Unmarshal([]byte(matchJSON), &rule.Match); err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
			continue
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// Update 更新重写规则
func (s *Store) Update(rule *RewriteRule) error {
	rule.UpdatedAt = time.Now()

	matchJSON, err := json.Marshal(rule.Match)
	if err != nil {
		return fmt.Errorf("序列化匹配条件失败: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	query := `
		UPDATE rewrite_rules
		SET name = ?, enabled = ?, priority = ?, match_config = ?, actions_config = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		rule.Name, rule.Enabled, rule.Priority,
		string(matchJSON), string(actionsJSON),
		rule.UpdatedAt, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("更新重写规则失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("重写规则不存在")
	}

	return nil
}

// Delete 删除重写规则
func (s *Store) Delete(id string) error {
	query := "DELETE FROM rewrite_rules WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除重写规则失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("重写规则不存在")
	}

	return nil
}

// Toggle 启用/禁用重写规则
func (s *Store) Toggle(id string) (bool, error) {
	rule, err := s.GetByID(id)
	if err != nil {
		return false, err
	}
	if rule == nil {
		return false, fmt.Errorf("重写规则不存在")
	}

	rule.Enabled = !rule.Enabled
	rule.UpdatedAt = time.Now()

	query := "UPDATE rewrite_rules SET enabled = ?, updated_at = ? WHERE id = ?"
	_, err = s.db.Exec(query, rule.Enabled, rule.UpdatedAt, id)
	if err != nil {
		return false, fmt.Errorf("切换重写规则状态失败: %w", err)
	}

	return rule.Enabled, nil
}

// GetEnabled 获取所有启用的重写规则
func (s *Store) GetEnabled() ([]*RewriteRule, error) {
	query := `
		SELECT id, name, enabled, priority, match_config, actions_config, created_at, updated_at
		FROM rewrite_rules
		WHERE enabled = TRUE
		ORDER BY priority ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*RewriteRule
	for rows.Next() {
		rule := &RewriteRule{}
		var matchJSON, actionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Enabled, &rule.Priority,
			&matchJSON, &actionsJSON,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(matchJSON), &rule.Match)
		json.Unmarshal([]byte(actionsJSON), &rule.Actions)

		rules = append(rules, rule)
	}

	return rules, nil
}

// Reorder 重新排序重写规则
func (s *Store) Reorder(ids []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for i, id := range ids {
		_, err := tx.Exec("UPDATE rewrite_rules SET priority = ? WHERE id = ?", i, id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
