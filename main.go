package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var fConfigFile = flag.String("c", "statsgod.config", "path to config file")
var fStatsPort = flag.Int("p", 16536, "stats listen port (UDP)")
var fHttpPort = flag.Int("h", 16530, "HTTP listen port")

func main() {
	flag.Parse()

	cf, err := ReadConfig(*fConfigFile)
	if err != nil {
		log.Fatalf("%v: error reading config file", err)
	}

	updateChan := make(chan StatUpdate, 10)
	funcChan := make(chan func(c *Core), 10)

	core := coreFromConfig(cf)

	go core.Loop(updateChan, funcChan)

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		DataHandler(w, r, funcChan)
	})
	go func() {
		err := http.ListenAndServe(":"+strconv.FormatInt(int64(*fHttpPort), 10), nil)
		if err != nil {
			log.Fatalf("%v: error starting web server", err)
		}
	}()

	udpAddress, err := net.ResolveUDPAddr("udp", ":"+strconv.FormatInt(int64(*fStatsPort), 10))
	if err != nil {
		log.Fatalf("%v: error resolving UDP port", err)
	}
	udpc, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalf("%v: error listening to UDP port", err)
	}

	go func() {
		b := make([]byte, 512)
		for {
			n, _, err := udpc.ReadFromUDP(b)
			if err != nil {
				log.Printf("%v: error reading udp packet", err)
			}
			if n > 0 {
				forwardPacket(string(b[:n]), updateChan)
			}
		}
	}()
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP)
	for {
		select {
		case _, ok := <-sigChan:
			if !ok {
				return
			}
			funcChan <- func(c *Core) {
				updateConfig(cf, c)
				cf.Write(*fConfigFile)
			}
		}
	}
}

func coreFromConfig(cf *ConfigFile) *Core {
	c := NewCore()
	for _, s := range cf.Stats {
		c.Stats[s.Key] = &StatRecord{
			Name:      s.Name,
			IsCounter: s.IsCounter,
			Capacity:  s.Capacity,
		}
	}
	return c
}

func updateConfig(cf *ConfigFile, c *Core) {
	for k, r := range c.Stats {
		found := false
		for _, s := range cf.Stats {
			if k == s.Key {
				found = true
			}
			break
		}
		if !found {
			cf.Stats = append(cf.Stats, StatConfig{Name: r.Name, Key: k, IsCounter: r.IsCounter})
		}
	}
}

func forwardPacket(s string, suc chan StatUpdate) {
	split := strings.Split(s, " ")
	if len(split) != 3 {
		log.Printf("%s: bad packet", s)
		return
	}
	f, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		log.Printf("%s: bad value (%s)", split[1], split[0])
		return
	}
	t, err := strconv.ParseInt(strings.TrimSpace(split[2]), 10, 64)
	if err != nil {
		log.Printf("%s: bad timestamp (%s)", split[2], split[0])
		return
	}
	suc <- StatUpdate{Key: split[0], Value: f, Timestamp: t}

}
