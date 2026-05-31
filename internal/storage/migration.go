package storage

import (
	"fmt"
	"log"
)

// Migration 数据库迁移
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations 迁移列表
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_traffic_table",
		SQL: `
		CREATE TABLE IF NOT EXISTS traffic (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			method TEXT NOT NULL,
			url TEXT NOT NULL,
			host TEXT,
			server_ip TEXT,
			content_type TEXT,
			status_code INTEGER,
			duration_ms INTEGER,
			req_headers TEXT,
			req_body BLOB,
			resp_headers TEXT,
			resp_body BLOB,
			raw_req BLOB,
			raw_resp BLOB,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_traffic_timestamp ON traffic(timestamp);
		CREATE INDEX IF NOT EXISTS idx_traffic_host ON traffic(host);
		`,
	},
	{
		Version: 2,
		Name:    "create_migration_history",
		SQL: `
		CREATE TABLE IF NOT EXISTS migration_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
	},
	{
		Version: 3,
		Name:    "create_breakpoints_table",
		SQL: `
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
		`,
	},
	{
		Version: 4,
		Name:    "create_rewrite_rules_table",
		SQL: `
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
		`,
	},
	{
		Version: 5,
		Name:    "create_collections_table",
		SQL: `
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
		`,
	},
	{
		Version: 6,
		Name:    "create_environments_table",
		SQL: `
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
		`,
	},
	{
		Version: 7,
		Name:    "create_settings_table",
		SQL: `
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
	},
}

// RunMigrations 执行数据库迁移
func (s *Storage) RunMigrations() error {
	// 先创建迁移历史表
	if _, err := s.DB.Exec(migrations[1].SQL); err != nil {
		return fmt.Errorf("创建迁移历史表失败: %w", err)
	}

	// 获取已应用的迁移版本
	applied, err := s.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}

	// 执行未应用的迁移
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		log.Printf("[INFO] 执行迁移: v%d - %s", m.Version, m.Name)

		// 在事务中执行迁移
		tx, err := s.DB.Begin()
		if err != nil {
			return fmt.Errorf("开始事务失败: %w", err)
		}

		if _, err := tx.Exec(m.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行迁移 SQL 失败: %w", err)
		}

		// 记录迁移历史
		if _, err := tx.Exec(
			"INSERT INTO migration_history (version, name) VALUES (?, ?)",
			m.Version, m.Name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("记录迁移历史失败: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}

		log.Printf("[INFO] 迁移 v%d 完成", m.Version)
	}

	return nil
}

// getAppliedMigrations 获取已应用的迁移版本
func (s *Storage) getAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := s.DB.Query("SELECT version FROM migration_history")
	if err != nil {
		// 表可能不存在，返回空 map
		return applied, nil
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("扫描迁移版本失败: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// GetMigrationStatus 获取迁移状态
func (s *Storage) GetMigrationStatus() ([]map[string]interface{}, error) {
	var status []map[string]interface{}

	for _, m := range migrations {
		applied := false
		var appliedAt string

		err := s.DB.QueryRow(
			"SELECT applied_at FROM migration_history WHERE version = ?",
			m.Version,
		).Scan(&appliedAt)

		if err == nil {
			applied = true
		}

		status = append(status, map[string]interface{}{
			"version":    m.Version,
			"name":       m.Name,
			"applied":    applied,
			"applied_at": appliedAt,
		})
	}

	return status, nil
}
