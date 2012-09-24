package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MetricUpdate struct {
	Name      string
	Value     float64
	Timestamp int64
	IsCounter bool
}

type metric struct {
	name      string
	value     float64
	updatedAt int64
	isCounter bool
	sentAt    int64
	mutex     sync.Mutex
}

func (m *metric) update(value float64, updated int64, isCounter bool) error {
	if isCounter != m.isCounter {
		return fmt.Errorf("%s: metric type mismatch", m.name)
	}
	m.mutex.Lock()
	if isCounter {
		m.value += value
	} else {
		m.value = value
	}
	m.updatedAt = updated
	m.mutex.Unlock()
	return nil
}

type Server struct {
	StatsPort    int64
	metrics      map[string]*metric
	metricsMutex sync.Mutex
	updateCond   *sync.Cond
	updateChan   chan MetricUpdate
	udpConn      *net.UDPConn
}

func NewServer(statsPort int64, updates chan MetricUpdate) *Server {
	s := &Server{
		StatsPort:  statsPort,
		metrics:    make(map[string]*metric),
		updateCond: sync.NewCond(new(sync.Mutex)),
		updateChan: updates,
	}

	return s
}

func (s *Server) Loop() error {

	go s.sendLoop()

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
			s.handlePacket(string(b[:n]))
		}
	}

	return nil
}

func (s *Server) sendLoop() {
	for {
		nowMillis := time.Now().UnixNano() / (1e6)

		s.metricsMutex.Lock()
		for _, m := range s.metrics {
			if m.sentAt < m.updatedAt && (nowMillis-m.sentAt) > 200 {
				s.updateChan <- MetricUpdate{Name: m.name, Value: m.value, Timestamp: (m.updatedAt / 1000), IsCounter: m.isCounter}
				m.sentAt = nowMillis
			}
		}
		s.metricsMutex.Unlock()

		// not exactly typical use of condition variables, but all we really want to do
		// is wait for an update to come in...
		s.updateCond.L.Lock()
		s.updateCond.Wait()
		s.updateCond.L.Unlock()
	}
}

func (s *Server) Stop() {
	close(s.updateChan)
	s.udpConn.Close()
}

func (s *Server) handlePacket(str string) {
	split := strings.Split(str, " ")
	if len(split) != 3 {
		log.Printf("%s: bad packet", str)
		return
	}
	name := split[0]
	f, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		log.Printf("%s: bad value (%s)", split[1], name)
		return
	}
	isCounter := true
	switch split[2] {
	case "G":
		isCounter = false
	case "C":
		isCounter = true
	default:
		log.Printf("%s: bad value (%s)", split[1], name)
		return
	}
	t := time.Now().UnixNano() / (1e6)

	m, ok := s.metrics[name]
	if !ok {
		m = &metric{name: name, value: f, updatedAt: t, isCounter: isCounter}
		s.metricsMutex.Lock()
		s.metrics[name] = m
		s.metricsMutex.Unlock()
	}
	err = m.update(f, t, isCounter)
	if err != nil {
		log.Printf("%v", err)
	}
	s.updateCond.Signal()
	return
}
