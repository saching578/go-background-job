# 🚀 Go Background Job Processing System (Kafka-based)

## 📌 Problem Statement

In real-world backend systems, some operations are **long-running** or **unreliable**, such as:

- generating reports  
- sending emails  
- processing large datasets  
- calling slow or flaky third-party APIs  

If these operations are handled **synchronously** inside an API request, it leads to:

❌ Slow APIs  
❌ Poor user experience  
❌ Difficult retry handling  
❌ Tight coupling between components  

---

## ✅ Solution Overview

This project implements a **production-style asynchronous background job processing system** using **Go, Kafka, and SQLite**.

### Core ideas:
- API responds immediately
- Background jobs are processed asynchronously
- Kafka is used for event-driven execution
- Jobs are retried automatically on failure
- No database polling
- Clear separation between API and worker

---

## 🏗 System Architecture

Client
|
| POST /jobs
v
API Service (Go)
|
| Store job in SQLite DB
| Publish job_id to Kafka
v
Kafka (job-events topic)
|
v
Worker Service (Go)
|
| Execute job
| Retry on failure
| Update job status
v
SQLite Database

## 🔁 Job Lifecycle

PENDING -> RUNNING -> DONE

-> FAILED (after max retries)

   - Failed jobs are retried automatically

   - Retry count is tracked in DB

   - Infinite retries are prevented



## ▶️ How to Run This Project Locally

✅ Prerequisites

Make sure the following are installed:

- **Go (>= 1.22)**
- **Docker Desktop**
- **Git**

---

1️⃣ Clone the Repository

git clone https://github.com/saching578/go-background-job.git
cd go-background-job

2️⃣ Start Kafka & Zookeeper

 - Ensure Docker Desktop is running, then:

 - docker compose up -d

Verify containers:

 - docker ps

You should see:

 - background-job-zookeeper-1

 - background-job-kafka-1


3️⃣ Start the API Service

 - go run ./api

Expected output:

 - API running on :8080

4️⃣ Start the Worker Service

Open a new terminal:

 - go run ./worker


Expected output:

 - Worker listening to Kafka topic: job-events

5️⃣ Create a Job (API Request)

Using PowerShell:

Invoke-RestMethod `
  -Uri http://localhost:8080/jobs `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"type":"report","payload":{"user_id":1}}'


Response:

job_id status
------ ------
1      PENDING

6️⃣ Observe Worker Logs

The worker terminal will show:

Job execution

Retry attempts (if failure occurs)

Final job status (DONE / FAILED)

7️⃣ Check Job Status
Invoke-RestMethod http://localhost:8080/jobs/1

🔁 Retry Mechanism

Jobs are retried automatically on failure

Retry limit is enforced

Retries are triggered via Kafka events

No polling or cron jobs are used

🚀 Why Kafka Is Used

Decouples API from worker

Enables horizontal scaling

Avoids database polling

Reliable message delivery

Event-driven architecture
