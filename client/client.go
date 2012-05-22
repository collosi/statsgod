package client

import (
	"fmt"
	"io"
	"net"
	"time"
)

type update struct {
	k string
	v float64
}

type Client struct {
	c          *net.UDPConn
	updateChan chan update
}

func Dial(addr string) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return nil, err
	}

	client := &Client{c, make(chan update, 10)}
	go func() {
		for u := range client.updateChan {
			client.forwardUpdate(u)
		}
	}()
	return client, nil
}

func (c *Client) Update(k string, value float64) {
	c.updateChan <- update{k, value}
}

func (c *Client) forwardUpdate(u update) {
	timestamp := time.Now().Unix()
	s := fmt.Sprintf("%s %f %d", u.k, u.v, timestamp)
	io.WriteString(c.c, s)
}
