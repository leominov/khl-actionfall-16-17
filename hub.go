package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Redis pool
	pool *redis.Pool

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func newHub() *Hub {
	return &Hub{
		pool:       newPool(),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// readPump pumps messages from the redis to the hub
func (h *Hub) readPump() {
	conn := h.pool.Get()

	defer conn.Close()

	if conn.Err() != nil {
		logrus.Fatalf("Redis pool error: %+v", conn.Err())
	}

	psc := redis.PubSubConn{conn}
	if err := psc.Subscribe("example"); err != nil {
		logrus.Fatalf("Redis subscribe error: %+v", err)
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			h.broadcast <- v.Data
		case error:
			logrus.Errorf("Redis connection refused: %+v", v)

			if conn.Err() == nil {
				break
			}

			conn = h.pool.Get()
			if conn.Err() != nil {
				logrus.Info("Redis: Wait for a 10 seconds")
				time.Sleep(10 * time.Second)
			} else {
				logrus.Info("Redis: connection established")
				psc = redis.PubSubConn{conn}
				psc.Subscribe("example")
			}
		}
	}
}

func (h *Hub) run() {
	var clientsLen int

	for {
		select {
		case client := <-h.register:
			logrus.Debugf("Register client: %s", client.conn.RemoteAddr())
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				logrus.Debugf("Unregister client: %s", client.conn.RemoteAddr())
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			clientsLen = len(h.clients)

			if clientsLen == 0 {
				logrus.Info("Skip broadcast message, no clients found")
				break
			}

			logrus.Info("Broadcast message")
			logrus.Infof("Clients: %d", clientsLen)
			logrus.Infof("Start time: %s", time.Now().UTC().String())
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			logrus.Infof("End time: %s", time.Now().UTC().String())
		}
	}
}
