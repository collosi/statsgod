package main

import (
	"bitbucket.org/kardianos/service"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/laslowh/statsgod/server"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var fConfigFile = flag.String("c", "/opt/statsgod/statsgod.config", "path to config file")
var fStatsPort = flag.Int64("p", 16536, "stats listen port (UDP)")
var fHttpPort = flag.Int64("h", 16530, "HTTP listen port")

func main() {
	flag.Parse()

	var displayName = "StatsGod statistics server"
	var desc = "Aggregates and serves statistics"
	var ws, err = service.NewService("statsgod", displayName, desc)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s unable to start: %s", displayName, err)
		return
	}
	if len(flag.Args()) > 0 {
		var err error
		verb := flag.Args()[0]
		switch verb {
		case "install":
			err = ws.Install()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: failed to install %q\n", err, displayName)
				return
			}
			ws.LogInfo("%q: service installed.", displayName)
		case "remove":
			err = ws.Remove()
			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
				return
			}
			ws.LogInfo("%q: service removed.", displayName)
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown command", verb)
		}
		return
	}
	err = ws.Run(func() error {
		// start
		err := doStart()
		if err != nil {
			return err
		}
		ws.LogInfo("%q: running", displayName)
		return nil
	}, func() error {
		// stop
		doStop()
		ws.LogInfo("%q: stopping", displayName)
		return nil
	})
	if err != nil {
		ws.LogError(err.Error())
	}
}

func doStart() error {
	cf, err := server.ReadConfig(*fConfigFile)
	if err != nil {
		return err
	}
	s := server.NewServer(cf, *fConfigFile, *fStatsPort)

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		dataHandler(w, r, s.FuncChan)
	})

	tcpAddress, err := net.ResolveTCPAddr("tcp", ":"+strconv.FormatInt(*fHttpPort, 10))
	if err != nil {
		return err
	}
	tcpConn, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		return err
	}
	go func() {
		http.Serve(tcpConn, nil)
	}()

	go s.Loop()
	return nil
}

func doStop() {

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
