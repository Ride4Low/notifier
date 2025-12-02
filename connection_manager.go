package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ride4Low/contracts/env"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type WSMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// connWrapper is a wrapper around the websocket connection to allow for thread-safe operations
// This is necessary because the websocket connection is not thread-safe
type connectionWrapper struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

type ConnectionManager struct {
	connections map[string]*connectionWrapper
	mutex       sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connectionWrapper),
		mutex:       sync.RWMutex{},
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Host)
		return r.Host == env.GetString("ALLOWED_ORIGIN", "localhost:8082") || r.Host == env.GetString("ALLOWED_ORIGIN", "localhost:3000")
	},
}

func (cm *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (cm *ConnectionManager) Add(id string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[id] = &connectionWrapper{conn: conn, mutex: sync.Mutex{}}
}

func (cm *ConnectionManager) Remove(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.connections, id)
}

func (cm *ConnectionManager) Get(id string) (*connectionWrapper, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	conn, ok := cm.connections[id]
	return conn, ok
}

func (cm *ConnectionManager) SendMessage(id string, msg WSMessage) error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	conn, ok := cm.connections[id]
	if !ok {
		return ErrConnectionNotFound
	}
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	return conn.conn.WriteJSON(msg)
}

func (cm *ConnectionManager) Broadcast(msg WSMessage) error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	for _, conn := range cm.connections {
		conn.mutex.Lock()
		defer conn.mutex.Unlock()
		if err := conn.conn.WriteJSON(msg); err != nil {
			return err
		}
	}
	return nil
}
