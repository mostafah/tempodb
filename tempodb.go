// Package tempodb provides an easy and asynchronous API for sending data points
// to TempoDB. It uses TempoDB's official Go package for implemenation:
// https://github.com/tempodb/tempodb-go.
//
// Send and SendId functions only put data points on a channel. Worker
// goroutines read from this channel and send data points to TempoDB.
package tempodb

// idea: retry after failures
// send multiple data points together

import (
	"github.com/tempodb/tempodb-go"
	"log"
	"sync"
	"time"
)

const (
	// size of buffer channel
	bufSize = 1000
	// number of workers
	workers = 3
)

// Type point holds a single data point as a time and value pair for a data
// series. The data series is defined either with the id or the key; if they key
// is set the id is empty, and vice versa.
type point struct {
	id, key string
	time    time.Time
	value   float64
}

// Type Client holds a connection to a TempoDB database.
type Client struct {
	c      *tempodb.Client
	points chan *point
	wg     sync.WaitGroup
	Debug  bool // set to true if you want errors to be logged
}

// NewClient creates a new client.
func NewClient(key, secret string) *Client {
	c := &Client{
		c:      tempodb.NewClient(key, secret),
		points: make(chan *point, bufSize),
	}
	for i := 0; i < workers; i++ {
		go c.worker()
	}
	return c
}

// Send posts a value to a series by its key.
func (c *Client) Send(key string, value float64) {
	c.points <- &point{key: key, time: time.Now(), value: value}
}

// SendId posts a value to a series by its id.
func (c *Client) SendId(id string, value float64) {
	c.points <- &point{id: id, time: time.Now(), value: value}
}

// worker sends each data point queued in channel to TempoDB.
func (c *Client) worker() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		p, ok := <-c.points
		if !ok {
			return // channel closed
		}
		c.send(p)
	}
}

// send actually sends a data point.
func (c *Client) send(p *point) {
	data := []*tempodb.DataPoint{{Ts: p.time, V: p.value}}
	var err error
	if p.id == "" {
		err = c.c.WriteKey(p.key, data)
	} else {
		err = c.c.WriteId(p.id, data)
	}
	if c.Debug && err != nil {
		log.Println("can't write data point:", err)
	}
}

// Finish waits for all the queued data points to be sent. Sending values after
// finishing will lead to a run-time panic.
func (c *Client) Finish(timeout time.Duration) bool {
	done := make(chan bool)
	go func() {
		close(c.points)
		c.wg.Wait()
		done <- true
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
