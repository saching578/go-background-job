package handler

import (
	"background-job/api/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
)

type CreateJobRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func CreateJob(db *sql.DB, producer sarama.SyncProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1️⃣ Allow only GET
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		jobID, err := repository.CreateJob(db, req.Type, req.Payload)
		if err != nil {
			http.Error(w, "failed to create job", http.StatusInternalServerError)
			return
		}

		// 🔔 Publish job_id to Kafka
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "job-events",
			Value: sarama.StringEncoder(strconv.FormatInt(jobID, 10)),
		})
		if err != nil {
			http.Error(w, "failed to publish job event", http.StatusInternalServerError)
			return
		}

		// 4️⃣ Return response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id": jobID,
			"status": "PENDING",
		})
	}
}

func GetJob(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// existing logic (unchanged)
	}
}
