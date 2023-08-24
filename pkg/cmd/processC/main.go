package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adjust/rmq/v5"
	alog "github.com/apex/log"
	japi "github.com/dcalliet/cscjapi"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/dcalliet/cscjapi/devtools"
	"github.com/redis/go-redis/v9"
)

func main() {
	errChan := make(chan error)
	// Load needed configuration for the Service
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
	for {
		if err = db.Ping(); err != nil {
			alog.WithError(err).Error("unable to establish a databse connection, retry in 5 seconds")
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			break
		}
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

	var connection rmq.Connection

	for {
		connection, err = rmq.OpenConnectionWithRedisOptions("my service", redis_options, errChan)
		if err != nil {
			alog.WithError(err).Error("failed to open redis connection, retry in 5 seconds.")
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			break
		}
	}

	taskQueue, err := connection.OpenQueue(japi.JOBS_QUEUE_NAME)
	if err != nil {
		alog.WithError(err).Error("failed to open queue")
		return
	}

	taskQueue.StartConsuming(1, time.Duration(application_config.EnvWorkerPollingWaitInSeconds)*time.Second)

	consumer := &japi.Consumer{
		Db: db,
	}
	_, err = taskQueue.AddConsumer("job-consumer", consumer)

	if err != nil {
		alog.WithError(err).Error("failed to have worker begin cron job.")
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	<-connection.StopAllConsuming() // wait for all Consume() calls to finish

}
