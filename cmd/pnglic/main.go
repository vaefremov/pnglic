/*
 * PANGEA License Manager
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/vaefremov/pnglic/config"
	"github.com/vaefremov/pnglic/pkg/chkexprd"
	sw "github.com/vaefremov/pnglic/pkg/openapi"
)

var configPath = flag.String("c", "./pnglic_config.yaml", "Path to config file")
var writeConfigToFile = flag.Bool("x", false, "Generate config file filled with default parameters")

func main() {
	flag.Parse()
	conf := config.NewConfig(*configPath)
	conf.Report()
	if *writeConfigToFile {
		if err := conf.Write(*configPath); err != nil {
			log.Printf("Warning: %s", err.Error())
		}
	}
	go chkexprd.RunExpiryNotifications(conf)
	router := sw.NewRouter(conf)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Port),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
	// log.Fatal(router.Run(fmt.Sprintf(":%d", conf.Port)))
}
