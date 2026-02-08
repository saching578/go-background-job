package repository

import "database/sql"

func CreateJob(db *sql.DB, jobType string, payload []byte) (int64, error) {

	result, err := db.Exec(
		`INSERT INTO jobs (type, payload, status)
		 VALUES (?, ?, 'PENDING')`,
		jobType,
		payload,
	)
	if err != nil {
		return 0, err
	}

	jobID, err := result.LastInsertId()
	return jobID, err
}

func GetJobStatus(db *sql.DB, jobID int64) (string, error) {
	var status string

	err := db.QueryRow(
		`SELECT status FROM jobs WHERE id = $1`,
		jobID,
	).Scan(&status)

	return status, err
}
