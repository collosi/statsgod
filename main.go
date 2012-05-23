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
	s := server.NewServer(cf, *fConfigFile, *fHttpPort, *fStatsPort)

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		DataHandler(w, r, s.FuncChan)
	})

	tcpAddress, err := net.ResolveTCPAddr("tcp", ":"+strconv.FormatInt(s.HttpPort, 10))
	if err != nil {
		return err
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
	s.Loop(sigChan)
}
