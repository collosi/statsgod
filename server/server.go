package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	AccessControlAllowOrigin = "*"
)

func DataHandler(w http.ResponseWriter, r *http.Request, coreChan chan func(c *Core)) {
	k := r.FormValue("k")
	if k == "" {
		http.Error(w, "'k' required", http.StatusBadRequest)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	coreChan <- func(c *Core) {
		defer wg.Done()
		h := w.Header()
		if AccessControlAllowOrigin != "" {
			h.Set("Access-Control-Allow-Origin", AccessControlAllowOrigin)
		}
		h.Set("Content-Type", "application/json")
		s, ok := c.Stats[k]
		if !ok {
			http.Error(w, "{}", http.StatusNotFound)
			return
		}
		var values []Datum
		s.CopyValues(&values)
		enc := json.NewEncoder(w)
		err := enc.Encode(values)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	wg.Wait()
}

type Server struct {
	StatsPort      int64
	Core           *Core
	ConfigFile     *ConfigFile
	ConfigFilePath string
	UpdateChan     chan StatUpdate
	FuncChan       chan func(c *Core)
	udpConn        *net.UDPConn
}

func NewServer(cf *ConfigFile, configFilePath string, statsPort int64) *Server {
	s := &Server{
		StatsPort:      statsPort,
		Core:           NewCore(),
		ConfigFile:     cf,
		ConfigFilePath: configFilePath,
		UpdateChan:     make(chan StatUpdate, 10),
		FuncChan:       make(chan func(c *Core), 10),
	}

	for _, stat := range cf.Stats {
		s.Core.Stats[stat.Key] = &StatRecord{
			Name:      stat.Name,
			IsCounter: stat.IsCounter,
			Capacity:  stat.Capacity,
		}
	}

	return s
}

func (s *Server) Loop() error {

	go s.Core.Loop(s.UpdateChan, s.FuncChan)

	udpAddress, err := net.ResolveUDPAddr("udp", ":"+strconv.FormatInt(s.StatsPort, 10))
	if err != nil {
		return err
	}
	s.udpConn, err = net.ListenUDP("udp", udpAddress)
	if err != nil {
		return err
	}
	b := make([]byte, 512)
	for {
		n, _, err := s.udpConn.ReadFromUDP(b)
		if err != nil {
			log.Printf("%v: error reading udp packet", err)
			break
		}
		if n > 0 {
			forwardPacket(string(b[:n]), s.UpdateChan)
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.udpConn.Close()
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
