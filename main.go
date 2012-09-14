package main

import (
	"bitbucket.org/kardianos/service"
	"flag"
	"fmt"
	statsgod "github.com/laslowh/statsgod/server"
	"github.com/laslowh/statsgod/datadog"
	"os"
)

var fStatsPort = flag.Int64("p", 16536, "stats listen port (UDP)")
var fDefaultHost = flag.String("h", "unknown", "default 'host' to send to Datadog")
var fDefaultDevice = flag.String("d", "unknown", "default 'device' to send to Datadog")
var fApiKey = flag.String("ak", "", "Datadog API key")
var fApplicationKey = flag.String("pk", "", "Datadog application key")

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
	updateChan := make(chan statsgod.MetricUpdate)

	ddc := datadog.NewClient(*fApiKey, *fApplicationKey)

	s := statsgod.NewServer(*fStatsPort, updateChan)
	err = ws.Run(func() error {
		// start
		err := doStart(s)
		if err != nil {
			return err
		}
		go forwardUpdates(ddc, updateChan)
		ws.LogInfo("%q: running", displayName)
		return nil
	}, func() error {
		// stop
		doStop(s)
		ws.LogInfo("%q: stopping", displayName)
		return nil
	})
	if err != nil {
		ws.LogError(err.Error())
	}
}

func doStart(s *statsgod.Server) error {

	go s.Loop()
	return nil
}

func doStop(s *statsgod.Server) {
	s.Stop()
}

func forwardUpdates(ddc *datadog.Client, ch chan statsgod.MetricUpdate) {
	for u := range ch {
	    fmt.Printf("forwarding update %v\n", u)
		ddc.SendMetricUpdate(u.Name, u.Value, u.Timestamp, u.IsCounter, *fDefaultHost, *fDefaultDevice, nil)
	}
}
