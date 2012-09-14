package client

import (
	"fmt"
	//"io"
	"net"
	"sync"
)

type Client struct {
	c     *net.UDPConn
	mutex sync.Mutex
}

type MetricType byte

const (
	GAUGE_TYPE   = MetricType('G')
	COUNTER_TYPE = MetricType('C')
)

func Dial(addr string, bufferSize int) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return nil, err
	}

	client := &Client{c: c}
	return client, nil
}

func (c *Client) UpdateFloat64(k string, v float64, t MetricType) error {
	_, err := fmt.Fprintf(c.c, "%s %f %c", k, v, t)
	return err
}

func (c *Client) UpdateInt(k string, v int, t MetricType) error {
	_, err := fmt.Fprintf(c.c, "%s %d %c", k, v, t)
	return err
}

func (c *Client) Close() {
	c.c.Close()
}
