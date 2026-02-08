package main

import (
	"database/sql"
	"log"
	"net/http"

	"background-job/api/handler"
	"background-job/api/kafka"

	_ "modernc.org/sqlite"
)

func main() {
	// Open SQLite database (file-based)
	db, err := sql.Open("sqlite", "file:db/jobs.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}

	// Kafka producer
	producer, err := kafka.NewProducer([]string{"127.0.0.1:9092"})
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Create table if not exists
	defer producer.Close()

	// Create table if not exists
	createJobsTable(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", handler.CreateJob(db, producer))
	mux.HandleFunc("/jobs/", handler.GetJob(db))

	log.Println("API running on :8080")
	http.ListenAndServe(":8080", mux)
}

func createJobsTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3,
		last_error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("failed to create jobs table:", err)
	}
}
