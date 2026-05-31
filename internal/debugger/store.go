package debugger

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store 断点存储
type Store struct {
	db *sql.DB
}

// NewStore 创建新的断点存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init 初始化断点表
func (s *Store) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS breakpoints (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255),
		enabled BOOLEAN DEFAULT TRUE,
		phase VARCHAR(20) DEFAULT 'request',
		match_config TEXT,
		action_config TEXT,
		hit_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_breakpoints_enabled ON breakpoints(enabled);

	CREATE TABLE IF NOT EXISTS breakpoint_sessions (
		id VARCHAR(36) PRIMARY KEY,
		breakpoint_id VARCHAR(36),
		transaction_id INTEGER,
		phase VARCHAR(20),
		status VARCHAR(20) DEFAULT 'paused',
		original_data TEXT,
		modified_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME,
		FOREIGN KEY (breakpoint_id) REFERENCES breakpoints(id)
	);
	CREATE INDEX IF NOT EXISTS idx_bp_sessions_status ON breakpoint_sessions(status);
	CREATE INDEX IF NOT EXISTS idx_bp_sessions_breakpoint ON breakpoint_sessions(breakpoint_id);
	`
	_, err := s.db.Exec(query)
	return err
}

// Create 创建断点
func (s *Store) Create(bp *Breakpoint) error {
	if bp.ID == "" {
		bp.ID = uuid.New().String()
	}

	now := time.Now()
	bp.CreatedAt = now
	bp.UpdatedAt = now

	matchJSON, err := json.Marshal(bp.Match)
	if err != nil {
		return fmt.Errorf("序列化匹配条件失败: %w", err)
	}

	actionJSON, err := json.Marshal(bp.Action)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	query := `
		INSERT INTO breakpoints (id, name, enabled, phase, match_config, action_config, hit_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		bp.ID, bp.Name, bp.Enabled, bp.Phase,
		string(matchJSON), string(actionJSON),
		bp.HitCount, bp.CreatedAt, bp.UpdatedAt,
	)

	return err
}

// GetByID 根据 ID 获取断点
func (s *Store) GetByID(id string) (*Breakpoint, error) {
	query := `
		SELECT id, name, enabled, phase, match_config, action_config, hit_count, created_at, updated_at
		FROM breakpoints WHERE id = ?
	`

	bp := &Breakpoint{}
	var matchJSON, actionJSON string

	err := s.db.QueryRow(query, id).Scan(
		&bp.ID, &bp.Name, &bp.Enabled, &bp.Phase,
		&matchJSON, &actionJSON,
		&bp.HitCount, &bp.CreatedAt, &bp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询断点失败: %w", err)
	}

	if err := json.Unmarshal([]byte(matchJSON), &bp.Match); err != nil {
		return nil, fmt.Errorf("解析匹配条件失败: %w", err)
	}

	if err := json.Unmarshal([]byte(actionJSON), &bp.Action); err != nil {
		return nil, fmt.Errorf("解析动作失败: %w", err)
	}

	return bp, nil
}

// List 获取断点列表
func (s *Store) List() ([]*Breakpoint, error) {
	query := `
		SELECT id, name, enabled, phase, match_config, action_config, hit_count, created_at, updated_at
		FROM breakpoints
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询断点列表失败: %w", err)
	}
	defer rows.Close()

	var breakpoints []*Breakpoint
	for rows.Next() {
		bp := &Breakpoint{}
		var matchJSON, actionJSON string

		err := rows.Scan(
			&bp.ID, &bp.Name, &bp.Enabled, &bp.Phase,
			&matchJSON, &actionJSON,
			&bp.HitCount, &bp.CreatedAt, &bp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描断点失败: %w", err)
		}

		if err := json.Unmarshal([]byte(matchJSON), &bp.Match); err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(actionJSON), &bp.Action); err != nil {
			continue
		}

		breakpoints = append(breakpoints, bp)
	}

	return breakpoints, nil
}

// Update 更新断点
func (s *Store) Update(bp *Breakpoint) error {
	bp.UpdatedAt = time.Now()

	matchJSON, err := json.Marshal(bp.Match)
	if err != nil {
		return fmt.Errorf("序列化匹配条件失败: %w", err)
	}

	actionJSON, err := json.Marshal(bp.Action)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	query := `
		UPDATE breakpoints
		SET name = ?, enabled = ?, phase = ?, match_config = ?, action_config = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		bp.Name, bp.Enabled, bp.Phase,
		string(matchJSON), string(actionJSON),
		bp.UpdatedAt, bp.ID,
	)
	if err != nil {
		return fmt.Errorf("更新断点失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("断点不存在")
	}

	return nil
}

// Delete 删除断点
func (s *Store) Delete(id string) error {
	// 先删除关联的会话
	_, err := s.db.Exec("DELETE FROM breakpoint_sessions WHERE breakpoint_id = ?", id)
	if err != nil {
		return fmt.Errorf("删除断点会话失败: %w", err)
	}

	query := "DELETE FROM breakpoints WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除断点失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("断点不存在")
	}

	return nil
}

// Toggle 启用/禁用断点
func (s *Store) Toggle(id string) (bool, error) {
	bp, err := s.GetByID(id)
	if err != nil {
		return false, err
	}
	if bp == nil {
		return false, fmt.Errorf("断点不存在")
	}

	bp.Enabled = !bp.Enabled
	bp.UpdatedAt = time.Now()

	query := "UPDATE breakpoints SET enabled = ?, updated_at = ? WHERE id = ?"
	_, err = s.db.Exec(query, bp.Enabled, bp.UpdatedAt, id)
	if err != nil {
		return false, fmt.Errorf("切换断点状态失败: %w", err)
	}

	return bp.Enabled, nil
}

// IncrementHitCount 增加命中次数
func (s *Store) IncrementHitCount(id string) error {
	query := "UPDATE breakpoints SET hit_count = hit_count + 1 WHERE id = ?"
	_, err := s.db.Exec(query, id)
	return err
}

// GetEnabled 获取所有启用的断点
func (s *Store) GetEnabled() ([]*Breakpoint, error) {
	query := `
		SELECT id, name, enabled, phase, match_config, action_config, hit_count, created_at, updated_at
		FROM breakpoints
		WHERE enabled = TRUE
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakpoints []*Breakpoint
	for rows.Next() {
		bp := &Breakpoint{}
		var matchJSON, actionJSON string

		err := rows.Scan(
			&bp.ID, &bp.Name, &bp.Enabled, &bp.Phase,
			&matchJSON, &actionJSON,
			&bp.HitCount, &bp.CreatedAt, &bp.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(matchJSON), &bp.Match)
		json.Unmarshal([]byte(actionJSON), &bp.Action)

		breakpoints = append(breakpoints, bp)
	}

	return breakpoints, nil
}

// CreateSession 创建断点会话
func (s *Store) CreateSession(session *BreakpointSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	session.CreatedAt = time.Now()
	session.Status = SessionStatusPaused

	originalJSON, err := json.Marshal(session.Original)
	if err != nil {
		return fmt.Errorf("序列化原始数据失败: %w", err)
	}

	query := `
		INSERT INTO breakpoint_sessions (id, breakpoint_id, transaction_id, phase, status, original_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		session.ID, session.BreakpointID, session.TransactionID,
		session.Phase, session.Status, string(originalJSON), session.CreatedAt,
	)

	return err
}

// GetSessionByID 根据 ID 获取会话
func (s *Store) GetSessionByID(id string) (*BreakpointSession, error) {
	query := `
		SELECT id, breakpoint_id, transaction_id, phase, status, original_data, modified_data, created_at, resolved_at
		FROM breakpoint_sessions WHERE id = ?
	`

	session := &BreakpointSession{}
	var originalJSON, modifiedJSON sql.NullString
	var resolvedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&session.ID, &session.BreakpointID, &session.TransactionID,
		&session.Phase, &session.Status,
		&originalJSON, &modifiedJSON,
		&session.CreatedAt, &resolvedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	if originalJSON.Valid {
		json.Unmarshal([]byte(originalJSON.String), &session.Original)
	}
	if modifiedJSON.Valid {
		json.Unmarshal([]byte(modifiedJSON.String), &session.Modified)
	}
	if resolvedAt.Valid {
		session.ResolvedAt = &resolvedAt.Time
	}

	return session, nil
}

// GetActiveSessions 获取活跃会话列表
func (s *Store) GetActiveSessions() ([]*BreakpointSession, error) {
	query := `
		SELECT id, breakpoint_id, transaction_id, phase, status, original_data, modified_data, created_at, resolved_at
		FROM breakpoint_sessions
		WHERE status = 'paused'
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询活跃会话失败: %w", err)
	}
	defer rows.Close()

	var sessions []*BreakpointSession
	for rows.Next() {
		session := &BreakpointSession{}
		var originalJSON, modifiedJSON sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&session.ID, &session.BreakpointID, &session.TransactionID,
			&session.Phase, &session.Status,
			&originalJSON, &modifiedJSON,
			&session.CreatedAt, &resolvedAt,
		)
		if err != nil {
			continue
		}

		if originalJSON.Valid {
			json.Unmarshal([]byte(originalJSON.String), &session.Original)
		}
		if modifiedJSON.Valid {
			json.Unmarshal([]byte(modifiedJSON.String), &session.Modified)
		}
		if resolvedAt.Valid {
			session.ResolvedAt = &resolvedAt.Time
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateSession 更新会话
func (s *Store) UpdateSession(session *BreakpointSession) error {
	originalJSON, err := json.Marshal(session.Original)
	if err != nil {
		return fmt.Errorf("序列化原始数据失败: %w", err)
	}

	var modifiedJSON []byte
	if session.Modified != nil {
		modifiedJSON, err = json.Marshal(session.Modified)
		if err != nil {
			return fmt.Errorf("序列化修改数据失败: %w", err)
		}
	}

	query := `
		UPDATE breakpoint_sessions
		SET status = ?, original_data = ?, modified_data = ?, resolved_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		session.Status, string(originalJSON), string(modifiedJSON),
		session.ResolvedAt, session.ID,
	)

	return err
}

// DeleteSession 删除会话
func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec("DELETE FROM breakpoint_sessions WHERE id = ?", id)
	return err
}
