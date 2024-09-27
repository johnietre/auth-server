package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
  addr := flag.String("addr", "127.0.0.1:4074", "Address to listen on")
  logFile := flag.String("log-file", "", "File to log to (empty means stderr)")
  flag.Parse()

  if *logFile != "" {
    f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
      log.Fatalln("error opening log file:", err)
    }
    log.SetOutput(f)
  }

  log.Fatalln("error running server:", http.ListenAndServe(*addr, nil))
}
