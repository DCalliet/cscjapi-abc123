package csc

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/adjust/rmq/v5"
	"github.com/stretchr/testify/suite"
)

type JobsTestSuite struct {
	suite.Suite
	cleanup   func()
	mock      sqlmock.Sqlmock
	db        *sql.DB
	queueConn rmq.TestConnection
}

func (suite *JobsTestSuite) SetupSuite() {
	db, mock, err := sqlmock.New()
	if err != nil {
		suite.T().Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	suite.cleanup = func() {
		db.Close()
	}

	suite.mock = mock
	suite.db = db

	suite.queueConn = rmq.NewTestConnection()
}

func (suite *JobsTestSuite) SetupTest() {
	suite.queueConn.Reset()
}

func (suite *JobsTestSuite) TeardownSuite() {
	suite.cleanup()
}

/* Execute Suites */
func TestJobsTestSuite(t *testing.T) {
	suite.Run(t, new(JobsTestSuite))
}

func (suite *JobsTestSuite) Test_GetJobs() {
	// Test non status supplied
	suite.mock.
		ExpectQuery(regexp.QuoteMeta(job_select_sql)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"data",
			"status",
			"created_at",
			"published_at",
			"started_at",
			"acknowledged_at",
			"updated_at",
			"deleted_at",
		}))

	_, err := GetJobs(suite.db, "")
	suite.NoError(err)
}

func (suite *JobsTestSuite) Test_GetJobs_ImproperStatus() {
	// Test non status supplied
	suite.mock.
		ExpectQuery(regexp.QuoteMeta(job_select_sql)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"data",
			"status",
			"created_at",
			"published_at",
			"started_at",
			"acknowledged_at",
			"updated_at",
			"deleted_at",
		}))

	_, err := GetJobs(suite.db, "exception")
	suite.NoError(err)
}

func (suite *JobsTestSuite) Test_GetJobs_Status() {
	status := "processed"
	suite.mock.
		ExpectQuery(regexp.QuoteMeta(fmt.Sprintf(`%s AND job.status = %s`, job_select_sql, status))).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"data",
			"status",
			"created_at",
			"published_at",
			"started_at",
			"acknowledged_at",
			"updated_at",
			"deleted_at",
		}))

	_, err := GetJobs(suite.db, status)
	suite.NoError(err)
}

func (suite *JobsTestSuite) Test_PrepareJob() {
	new_job_id := 1
	test := Task{Message: "hello!"}
	b, err := json.Marshal(test)
	if !suite.NoError(err) {
		return
	}

	// Prepare Mocked DB for upcoming queries
	suite.mock.
		ExpectQuery(regexp.QuoteMeta(job_insert_sql)).
		WithArgs(string(b), "created", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(new_job_id))

	suite.mock.
		ExpectExec(regexp.QuoteMeta(fmt.Sprint(job_update_sql, ", published_at = $2 WHERE id = $3"))).
		WithArgs("published", sqlmock.AnyArg(), new_job_id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = PrepareJob(suite.db, suite.queueConn, b)

	suite.NoError(err)

	// Assert task delivered as expected
	h := make(http.Header)
	h.Set("job-id", fmt.Sprint(new_job_id))
	suite.Equal(string(rmq.PayloadBytesWithHeader(b, h)), suite.queueConn.GetDelivery(JOBS_QUEUE_NAME, 0))
}

func (suite *JobsTestSuite) Test_ProcessJob() {
	new_job_id := 1
	test := Task{Message: "hello!"}
	b, err := json.Marshal(test)
	if !suite.NoError(err) {
		return
	}
	h := make(http.Header)
	h.Set("job-id", fmt.Sprint(new_job_id))
	delivery := rmq.NewTestDeliveryString(string(rmq.PayloadBytesWithHeader(b, h)))

	// Prepare Mocked Database for expected calls
	suite.mock.
		ExpectExec(regexp.QuoteMeta(fmt.Sprint(job_update_sql, ", started_at = $2 WHERE id = $3"))).
		WithArgs("unacked", sqlmock.AnyArg(), new_job_id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	suite.mock.
		ExpectExec(regexp.QuoteMeta(fmt.Sprint(job_update_sql, ", acknowledged_at = $2 WHERE id = $3"))).
		WithArgs("processed", sqlmock.AnyArg(), new_job_id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = ProcessJob(suite.db, delivery)
	suite.NoError(err)
	suite.Equal(rmq.Acked, delivery.State)
}

func (suite *JobsTestSuite) Test_ProcessJob_RejectEmptyPayload() {
	new_job_id := 1
	h := make(http.Header)
	h.Set("job-id", fmt.Sprint(new_job_id))
	delivery := rmq.NewTestDeliveryString(rmq.PayloadWithHeader("", h))

	// Prepare Mocked Database for expected calls
	suite.mock.
		ExpectExec(regexp.QuoteMeta(fmt.Sprint(job_update_sql, ", started_at = $2 WHERE id = $3"))).
		WithArgs("unacked", sqlmock.AnyArg(), new_job_id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	suite.mock.
		ExpectExec(regexp.QuoteMeta(fmt.Sprint(job_update_sql, ", acknowledged_at = $2 WHERE id = $3"))).
		WithArgs("processed", sqlmock.AnyArg(), new_job_id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := ProcessJob(suite.db, delivery)
	suite.NoError(err)
	suite.Equal(rmq.Rejected, delivery.State)
}
