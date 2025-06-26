package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// TrafficData encapsulates all the information about a captured request/response pair.
type TrafficData struct {
	Method      string
	URL         string
	ServerIP    string
	ContentType string
	DurationMs  int64
	ReqHeaders  string
	ReqBody     []byte
	Status      string
	RespHeaders string
	RespBody    []byte
	RawReq      []byte
	RawResp     []byte
}

// Storage handles database operations.
type Storage struct {
	DB *sql.DB
}

// NewStorage creates a new connection to the database.
func NewStorage(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &Storage{DB: db}, nil
}

// InitTable creates the necessary table if it doesn't exist.
func (s *Storage) InitTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS traffic (
		id INT AUTO_INCREMENT PRIMARY KEY,
		method VARCHAR(10) NOT NULL,
		url TEXT NOT NULL,
		server_ip VARCHAR(45),
		content_type VARCHAR(255),
		duration_ms INT,
		request_headers TEXT,
		request_body LONGBLOB,
		response_status VARCHAR(50),
		response_headers TEXT,
		response_body LONGBLOB,
		raw_request LONGBLOB,
		raw_response LONGBLOB,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	log.Println("Database table 'traffic' is ready.")
	return nil
}

// SaveTraffic saves the captured traffic data to the database.
func (s *Storage) SaveTraffic(data *TrafficData) error {
	query := `
	INSERT INTO traffic (method, url, server_ip, content_type, duration_ms, request_headers, request_body, response_status, response_headers, response_body, raw_request, raw_response)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := s.DB.Exec(query, data.Method, data.URL, data.ServerIP, data.ContentType, data.DurationMs, data.ReqHeaders, data.ReqBody, data.Status, data.RespHeaders, data.RespBody, data.RawReq, data.RawResp)
	return err
}

// Close closes the database connection.
func (s *Storage) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}