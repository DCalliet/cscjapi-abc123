package main

import (
	"fmt"
	"net/http"
	"os"

	alog "github.com/apex/log"
	"github.com/dcalliet/cscjapi/devtools"
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

	// Construct Route(s) and attach Route Handler

	router := http.NewServeMux()

	router.HandleFunc("/v1/jobs", func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("ok"))
	})

	// Start Application

	if err := http.ListenAndServe(fmt.Sprint(":", application_config.EnvHTTPPort), router); err != nil {
		alog.WithError(err).Error(fmt.Sprint("unable to run server on port :", application_config.EnvHTTPPort))
	}
}
