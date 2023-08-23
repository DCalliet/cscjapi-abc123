package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/adjust/rmq/v5"
	alog "github.com/apex/log"
	japi "github.com/dcalliet/cscjapi"
	"github.com/dcalliet/cscjapi/devtools"
	"github.com/golang/gddo/httputil/header"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load needed configuration for the Service
	errChan := make(chan error)
	v, err := devtools.LoadConfig()
	if err != nil {
		alog.WithError(err).Error("failed to load needed configuration")
		os.Exit(1)
	}

	var application_config devtools.ApplicationConfig

	v.Unmarshal(&application_config)

	// Open Connections
	connstring := fmt.Sprint("postgresql://", application_config.EnvDBUsername, ":", application_config.EnvDBPassword, "@", application_config.EnvDBHostname, ":", application_config.EnvDBPort, "/", application_config.EnvDBName)
	db, err := sql.Open("pgx", connstring)
	if err != nil {
		alog.WithError(err).Error("failed to open database connection")
		os.Exit(1)
	}
	_, err = db.Query("SELECT 1;")
	if err != nil {
		alog.WithError(err).Error("failed to confirm db connection")
		os.Exit(1)
	}
	redis_options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", application_config.EnvRedisHostname, application_config.EnvRedisPort),
		Password: application_config.EnvDBPassword, // no password set
		DB:       0,                                // use default DB
	}
	connection, err := rmq.OpenConnectionWithRedisOptions("my service", redis_options, errChan)
	if err != nil {
		alog.WithError(err).Error("failed to open redis connection")
		os.Exit(1)
	}

	// Construct Route(s) and attach Route Handler

	router := http.NewServeMux()

	router.HandleFunc("/v1/jobs", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" || request.Method == "" {
			values := request.URL.Query()
			status := values.Get("status")
			jobs := []map[string]string{}
			rows, err := japi.GetJobs(db, status)
			if err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
				return
			}
			for rows.Next() {
				var (
					id              int
					data            string
					status          string
					created_at      string
					published_at    string
					started_at      string
					acknowledged_at string
					rejected_at     string
					updated_at      string
					deleted_at      string
				)
				err = rows.Scan(&id, &data, &status, &created_at, &published_at, &started_at, &acknowledged_at, &rejected_at, &updated_at, &deleted_at)
				if err != nil {
					break
				}
				found := map[string]string{
					"id":              fmt.Sprint(id),
					"data":            data,
					"status":          status,
					"created_at":      created_at,
					"published_at":    published_at,
					"started_at":      started_at,
					"acknowledged_at": acknowledged_at,
					"rejected_at":     rejected_at,
					"updated_at":      updated_at,
					"deleted_at":      deleted_at,
				}
				alog.WithField("found", found).Info("moving on")
				jobs = append(jobs, found)
			}
			if closeErr := rows.Close(); closeErr != nil {
				http.Error(response, closeErr.Error(), http.StatusInternalServerError)
				return
			}
			b, err := json.Marshal(jobs)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = response.Write(b)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if request.Method == "POST" {
			if request.Header.Get("Content-Type") != "" {
				value, _ := header.ParseValueAndParams(request.Header, "Content-Type")
				if value != "application/json" {
					msg := "Content-Type header is not application/json"
					http.Error(response, msg, http.StatusUnsupportedMediaType)
					return
				}
			}
			b, err := io.ReadAll(request.Body)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = japi.PrepareJob(db, connection, b)
			if err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
				return
			}
			response.WriteHeader(http.StatusNoContent)
			return
		} else {
			response.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Start Application
	alog.WithField("port", application_config.EnvHTTPPort).Info("start serverA")
	if err := http.ListenAndServe(fmt.Sprint(":", application_config.EnvHTTPPort), router); err != nil {
		alog.WithError(err).Error(fmt.Sprint("unable to run server on port :", application_config.EnvHTTPPort))
	}
}
