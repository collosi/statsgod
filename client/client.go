package client

import (
	"fmt"
	//"io"
	"net"
	"sync"
	"time"
)

type update struct {
	k string
	v float64
}

type Client struct {
	c          *net.UDPConn
	updateChan chan update
	closedSem  sync.WaitGroup
}

func Dial(addr string, bufferSize int) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return nil, err
	}

	client := &Client{c: c, updateChan: make(chan update, bufferSize)}
	client.closedSem.Add(1)
	go func() {
		for u := range client.updateChan {
			client.forwardUpdate(u)
		}
		c.Close()
		client.closedSem.Done()
	}()
	return client, nil
}

func (c *Client) Update(k string, value float64) {
	c.updateChan <- update{k, value}
}

func (c *Client) Close() {
	close(c.updateChan)
	c.closedSem.Wait()
}

func (c *Client) forwardUpdate(u update) {
	now := time.Now()
	timestamp := now.Unix()
	s := fmt.Sprintf("%s %f %d", u.k, u.v, timestamp)
	c.c.Write([]byte(s))
}
