package main

import (
	"fmt"
	"os"
	"time"

	alog "github.com/apex/log"
	"github.com/dcalliet/cscjapi/devtools"
	"github.com/go-co-op/gocron"
)

func main() {
	// Load needed configuration for the Service
	v, err := devtools.LoadConfig()
	if err != nil {
		alog.WithError(err).Error("failed to load needed configuration")
		os.Exit(1)
	}

	var application_config devtools.ApplicationConfig

	v.Unmarshal(&application_config)

	// Instance of GoCron Scheduler
	s := gocron.NewScheduler(time.UTC)
	alog.Info(fmt.Sprintf("Cron Schedule '%s'", application_config.EnvWorkerCronSchedule))
	_, err = s.CronWithSeconds(application_config.EnvWorkerCronSchedule).Do(func() {
		alog.Info("ok")
	})
	if err != nil {
		alog.WithError(err).Error("failed to have worker begin cron job.")
		os.Exit(1)
	}

	alog.Info("process C incrementally working on Queue Tasks")
	s.StartBlocking()
}
