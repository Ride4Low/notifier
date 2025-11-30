package main

import (
	"log"
	"net/http"
)

type Handler struct {
	cm *ConnectionManager
}

func newHandler(cm *ConnectionManager) *Handler {
	return &Handler{cm: cm}
}

func (h *Handler) handleRiders(w http.ResponseWriter, r *http.Request) {
	conn, err := h.cm.Upgrade(w, r)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	h.cm.Add(userID, conn)
	defer h.cm.Remove(userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Println("Received message:", string(message))
	}

}
