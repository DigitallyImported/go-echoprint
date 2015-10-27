package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/AudioAddict/go-echoprint/echoprint"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()
	defer glog.Flush()

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/debug", debugHandler).Methods("GET", "POST")
	router.HandleFunc("/query", queryHandler).Methods("GET", "POST")
	router.HandleFunc("/ingest", ingestHandler).Methods("POST")

	router.HandleFunc("/purge", purgeHandler).Methods("GET")

	loggingHandler := NewLoggingHandler(router)
	serverAddr := fmt.Sprintf(":%d", 8080)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: loggingHandler,
	}

	if err := echoprint.DBConnect(); err != nil {
		glog.Fatal(err)
	}
	defer echoprint.DBDisconnect()

	// TODO: gracefully stop http server (github.com/tylerb/graceful etc)
	glog.Infof("Starting server [%s]", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		glog.Fatal(err)
	}
}
