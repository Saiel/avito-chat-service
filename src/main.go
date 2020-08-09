package main

import (
	"bufio"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"

	_ "github.com/lib/pq"
)

func initErrorLogger() *log.Logger {
	errorFile, err := os.Create("error.log")
	if err != nil {
		log.Fatalln(err)
	}
	logWriter := bufio.NewWriter(errorFile)
	return log.New(logWriter, "", 0)
}

func main() {
	appConf := new(appEnvSettings)
	err := envconfig.Process("app", appConf)
	if err != nil {
		log.Fatalln(err)
	}

	handler := &Handler{}
	initDB(handler)
	mux := initMux(handler)
	errorLogger := initErrorLogger()
	server := &http.Server{
		Addr:     ":" + appConf.ServerPort,
		ErrorLog: errorLogger,
		Handler:  mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Cannot serve server: %v", err)
	}
}
