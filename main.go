package main

import (
	"flag"
	"github.com/laslowh/statsgod/server"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"encoding/json"
	"sync"
	"strconv"
)

var fConfigFile = flag.String("c", "statsgod.config", "path to config file")
var fStatsPort = flag.Int64("p", 16536, "stats listen port (UDP)")
var fHttpPort = flag.Int64("h", 16530, "HTTP listen port")

func main() {
	flag.Parse()

	cf, err := server.ReadConfig(*fConfigFile)
	if err != nil {
		log.Fatalf("%v: error reading config file", err)
	}
	s := server.NewServer(cf, *fConfigFile, *fStatsPort)

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		dataHandler(w, r, s.FuncChan)
	})

	tcpAddress, err := net.ResolveTCPAddr("tcp", ":"+strconv.FormatInt(*fHttpPort, 10))
	if err != nil {
		log.Fatalf("%v: error resolving http listen port %d", err, *fHttpPort) 
	}
	tcpConn, err := net.ListenTCP("tcp", tcpAddress)
	go func() {
		err := http.Serve(tcpConn, nil)
		if err != nil {
			return
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP)
	err = s.Loop(sigChan)
	if err != nil {
	   log.Printf("%v: error from loop", err)
	}
}

func dataHandler(w http.ResponseWriter, r *http.Request, coreChan chan func(c *server.Core)) {
	k := r.FormValue("k")
	if k == "" {
		http.Error(w, "'k' required", http.StatusBadRequest)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	coreChan <- func(c *server.Core) {
		defer wg.Done()
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Content-Type", "application/json")
		s, ok := c.Stats[k]
		if !ok {
			http.Error(w, "{}", http.StatusNotFound)
			return
		}
		var values []server.Datum
		s.CopyValues(&values)
		enc := json.NewEncoder(w)
		err := enc.Encode(values)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	wg.Wait()
}

