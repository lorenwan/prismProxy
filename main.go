package main

import (
	"log"
	"net/http"
	"web_test/proxy"
	"web_test/storage"
)

func main() {
	// Database connection string
	// Format: user:password@tcp(host:port)/dbname
	dsn := "root:my-secret-pw@tcp(127.0.0.1:3306)/proxy_data?charset=utf8mb4&parseTime=True"

	// Create a new storage instance
	db, err := storage.NewStorage(dsn)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the database table
	if err := db.InitTable(); err != nil {
		log.Fatalf("[FATAL] Failed to initialize database table: %v", err)
	}

	// Create a new proxy instance, passing the storage
	proxy, err := proxy.NewProxy(db)
	if err != nil {
		log.Fatalf("[FATAL] Failed to create proxy: %v", err)
	}

	// Create and start the HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	log.Println("Starting proxy server on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Could not listen on :8080: %v\n", err)
	}
}
