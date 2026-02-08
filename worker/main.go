package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	_ "modernc.org/sqlite"
)

const maxRetries = 3

func main() {
	// ---------------------------
	// Open SQLite DB
	// ---------------------------
	db, err := sql.Open("sqlite", "file:db/jobs.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ---------------------------
	// Kafka Producer (for retries)
	// ---------------------------
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// ---------------------------
	// Kafka Consumer
	// ---------------------------
	consumer, err := sarama.NewConsumer([]string{"127.0.0.1:9092"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(
		"job-events",
		0,
		sarama.OffsetNewest,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	log.Println("Worker listening to Kafka topic: job-events")

	// ---------------------------
	// Consume messages
	// ---------------------------
	for msg := range partitionConsumer.Messages() {
		jobID, err := strconv.ParseInt(string(msg.Value), 10, 64)
		if err != nil {
			log.Println("invalid job id:", err)
			continue
		}

		processJob(db, producer, jobID)
	}
}

// --------------------------------------------------
// JOB PROCESSING
// --------------------------------------------------

func processJob(db *sql.DB, producer sarama.SyncProducer, jobID int64) {
	log.Println("Processing job:", jobID)

	_, err := db.Exec(
		`UPDATE jobs SET status = 'RUNNING' WHERE id = ?`,
		jobID,
	)
	if err != nil {
		log.Println("failed to mark RUNNING:", err)
		return
	}

	err = doActualWork(jobID)
	if err != nil {
		handleFailure(db, producer, jobID, err)
		return
	}

	_, err = db.Exec(
		`UPDATE jobs SET status = 'DONE' WHERE id = ?`,
		jobID,
	)
	if err != nil {
		log.Println("failed to mark DONE:", err)
		return
	}

	log.Println("Job completed:", jobID)
}

// --------------------------------------------------
// SIMULATED WORK
// --------------------------------------------------

func doActualWork(jobID int64) error {
	// Simulate failure for learning
	if jobID%2 == 1 {
		return errors.New("simulated failure")
	}

	time.Sleep(3 * time.Second)
	return nil
}

// --------------------------------------------------
// FAILURE + RETRY HANDLING
// --------------------------------------------------

func handleFailure(
	db *sql.DB,
	producer sarama.SyncProducer,
	jobID int64,
	err error,
) {
	var retries int

	_ = db.QueryRow(
		`SELECT retry_count FROM jobs WHERE id = ?`,
		jobID,
	).Scan(&retries)

	retries++

	if retries >= maxRetries {
		_, _ = db.Exec(
			`UPDATE jobs 
			 SET status = 'FAILED', retry_count = ?, last_error = ?
			 WHERE id = ?`,
			retries, err.Error(), jobID,
		)

		log.Println("Job permanently FAILED:", jobID)
		return
	}

	_, _ = db.Exec(
		`UPDATE jobs 
		 SET status = 'PENDING', retry_count = ?, last_error = ?
		 WHERE id = ?`,
		retries, err.Error(), jobID,
	)

	// 🔁 Re-publish job to Kafka
	_, _, pubErr := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "job-events",
		Value: sarama.StringEncoder(strconv.FormatInt(jobID, 10)),
	})
	if pubErr != nil {
		log.Println("failed to re-publish job:", pubErr)
		return
	}

	log.Println("Retry scheduled for job:", jobID, "attempt:", retries)
}
