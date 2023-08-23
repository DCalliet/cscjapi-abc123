package csc

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
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
	alog.WithField("status", status).Info("GetJobs")
	defer alog.Trace("GetJobs").Stop(nil)
	validStatus := []string{
		"created",   // task not in queue
		"published", // task published to queue
		"unacked",   // task picked for consumption
		"rejected",  // task rejected
		"processed", // task completed
	}

	for _, testStatus := range validStatus {
		if testStatus == status {
			return db.Query(fmt.Sprintf(`%s AND job.status = %s`, job_select_sql, status))
		}
	}
	return db.Query(job_select_sql)
}

func PrepareJob(db *sql.DB, conn rmq.Connection, b []byte) (result sql.Result, err error) {
	alog.WithField("b", string(b)).Info("PrepareJob")
	defer alog.Trace("PrepareJob").Stop(nil)
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
	payload := rmq.PayloadBytesWithHeader(b, h)
	alog.Info(string(payload))
	err = queue.PublishBytes(payload)
	if err != nil {
		return
	}

	result, err = db.Exec(fmt.Sprint(job_update_sql, ", published_at = $2 WHERE id = $3"), "published", time.Now().Format(time.RFC3339), id)
	return
}

type Consumer struct {
	Db *sql.DB
}

func (c *Consumer) Consume(delivery rmq.Delivery) {
	ProcessJob(c.Db, delivery)
}

func ProcessJob(db *sql.DB, delivery rmq.Delivery) (result sql.Result, err error) {
	var id int
	var payload string
	if wh, ok := delivery.(rmq.WithHeader); ok {
		idToConv := wh.Header().Get("job-id")
		id, err = strconv.Atoi(idToConv)
		if err != nil {
			return
		}
		payload = delivery.Payload()
	} else {
		var header http.Header
		header, payload, err = rmq.ExtractHeaderAndPayload(delivery.Payload())
		if err != nil {
			return
		}
		idToConv := header.Get("job-id")
		id, err = strconv.Atoi(idToConv)
		if err != nil {
			return
		}
	}

	if err != nil {
		alog.Info("Uh oh! Missing JobId. Rejecting")
		err = delivery.Reject()
		result, _ = db.Exec(fmt.Sprint(job_update_sql, ", rejected_at = $2 WHERE id = $3"), "rejected", time.Now().Format(time.RFC3339), id)
		return
	}
	result, err = db.Exec(fmt.Sprint(job_update_sql, ", started_at = $2 WHERE id = $3"), "unacked", time.Now().Format(time.RFC3339), id)
	if err != nil {
		return
	}
	alog.WithField("Payload", payload).Info("Processed Queue Item.")
	if payload == "" {
		alog.Info("Uh oh! Empty Payload. Rejecting")
		err = delivery.Reject()
		result, _ = db.Exec(fmt.Sprint(job_update_sql, ", rejected_at = $2 WHERE id = $3"), "rejected", time.Now().Format(time.RFC3339), id)
		return
	} else {
		err = delivery.Ack()
		result, _ = db.Exec(fmt.Sprint(job_update_sql, ", acknowledged_at = $2 WHERE id = $3"), "processed", time.Now().Format(time.RFC3339), id)
		return
	}
}
