package csc

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/adjust/rmq/v5"
	alog "github.com/apex/log"
)

//go:embed job_select.sql
var job_select_sql string

//go:embed job_insert.sql
var job_insert_sql string

//go:embed job_update.sql
var job_update_sql string

var (
	JOBS_QUEUE_NAME = "jobs"
)

type Task struct {
	Message string `json:"message"`
}

func GetJobs(db *sql.DB, status string) (*sql.Rows, error) {
	validStatus := []string{
		"created",
		"published",
		"unacked",
		"rejected",
		"processed",
	}

	for _, testStatus := range validStatus {
		if testStatus == status {
			return db.Query(fmt.Sprintf(`%s AND "jobs"."status" = %s`, job_select_sql, status))
		}
	}
	return db.Query(job_select_sql)
}

func PrepareJob(db *sql.DB, conn rmq.Connection, b []byte) (result sql.Result, err error) {
	rows, err := db.Query(job_insert_sql, string(b), "created", time.Now().Format(time.RFC3339))
	if err != nil {
		return
	}
	var id int
	rows.Next()
	if err = rows.Scan(&id); err != nil {
		return
	}
	if err = rows.Close(); err != nil {
		return
	}

	queue, err := conn.OpenQueue(JOBS_QUEUE_NAME)
	if err != nil {
		return
	}

	h := make(http.Header)
	h.Set("job-id", fmt.Sprint(id))
	err = queue.PublishBytes(rmq.PayloadBytesWithHeader(b, h))
	if err != nil {
		return
	}

	result, err = db.Exec(fmt.Sprint(job_update_sql, ", published_at = ? WHERE id = ?"), "published", time.Now().Format(time.RFC3339), id)
	return
}

func ProcessJob(db *sql.DB, delivery rmq.Delivery) (result sql.Result, err error) {
	header, payload, err := rmq.ExtractHeaderAndPayload(delivery.Payload())
	if err != nil {
		return
	}
	id := header.Get("job-id")

	result, err = db.Exec(fmt.Sprint(job_update_sql, ", started_at = ? WHERE id = ?"), "unacked", time.Now().Format(time.RFC3339), id)
	if err != nil {
		return
	}
	alog.WithField("Payload", payload).Info("Processed Queue Item.")
	if payload == "" {
		alog.Info("Uh oh! Empty Payload. Rejecting")
		err = delivery.Reject()
		result, _ = db.Exec(fmt.Sprint(job_update_sql, ", rejected_at = ? WHERE id = ?"), "rejected", time.Now().Format(time.RFC3339), id)
		return
	} else {
		err = delivery.Ack()
		result, _ = db.Exec(fmt.Sprint(job_update_sql, ", acknowledged_at = ? WHERE id = ?"), "processed", time.Now().Format(time.RFC3339), id)
		return
	}
}
