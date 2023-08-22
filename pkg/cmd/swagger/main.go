package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	alog "github.com/apex/log"
	"github.com/dcalliet/cscjapi/devtools"
	"github.com/flowchartsman/swaggerui"
)

func main() {
	// Load needed configuration for the Service

	v, err := devtools.LoadConfig()
	if err != nil {
		alog.WithError(err).Error("failed to load needed configuration")
	}

	var config devtools.SwaggerConfig

	v.Unmarshal(&config)

	// Load swagger file
	target := filepath.Join(config.EnvSwaggerPath, config.EnvSwaggerFilename)
	spec, err := os.ReadFile(target)
	if err != nil {
		alog.WithError(err).Error(fmt.Sprintf("failed to load swagger file from %s", target))
	}

	// Construct Route(s) and attach Route Handler
	router := http.NewServeMux()

	router.Handle("/swagger/", http.StripPrefix("/swagger", swaggerui.Handler(spec)))

	// Start Application

	if err := http.ListenAndServe(fmt.Sprint(":", config.EnvHTTPPort), router); err != nil {
		alog.WithError(err).Error(fmt.Sprint("unable to run server on port :", config.EnvHTTPPort))
	}
}
