package main

import (
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
	pb "github.com/ride4Low/contracts/proto/driver"
)

type Handler struct {
	cm        *ConnectionManager
	ds        *DriverServiceClient
	publisher *rabbitmq.Publisher
}

func newHandler(cm *ConnectionManager, ds *DriverServiceClient, publisher *rabbitmq.Publisher) *Handler {
	return &Handler{cm: cm, ds: ds, publisher: publisher}
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

func (h *Handler) handleDrivers(w http.ResponseWriter, r *http.Request) {
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

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Println("No package slug provided")
		return
	}

	h.cm.Add(userID, conn)
	defer h.cm.Remove(userID)

	ctx := r.Context()

	driverReq := &pb.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	}

	// driverData, err := h.registerWithLog(ctx, driverReq)
	driverData, err := h.ds.client.RegisterDriver(ctx, driverReq)
	if err != nil {
		log.Println("Error registering driver:", err)
		return
	}

	if err := h.cm.SendMessage(userID, WSMessage{
		Type: events.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	defer func() {
		_, err := h.ds.client.UnregisterDriver(ctx, driverReq)
		if err != nil {
			log.Println("Error unregistering driver:", err)
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Println("Received message:", string(message))

		var msg WSDriverMessage
		if err := sonic.Unmarshal(message, &msg); err != nil {
			log.Println("Error unmarshalling message:", err)
			break
		}

		switch msg.Type {
		case events.DriverCmdLocation:
			// Handle driver location update in the future
			continue
		case events.DriverCmdTripAccept, events.DriverCmdTripDecline:
			if err := h.publisher.PublishMessage(ctx, msg.Type, events.AmqpMessage{
				OwnerID: userID,
				Data:    msg.Data,
			}); err != nil {
				log.Printf("Error publishing message: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// func (h *Handler) registerWithLog(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
// 	state := h.ds.GetState()
// 	log.Println(state)

// 	quitChan := make(chan struct{}, 1)
// 	defer func() {
// 		quitChan <- struct{}{}
// 	}()

// 	go func() {
// 		for {
// 			time.Sleep(time.Millisecond)

// 			select {
// 			case <-quitChan:
// 				return
// 			default:
// 				state := h.ds.GetState()
// 				log.Println(state)

// 			}
// 		}
// 	}()

// 	return h.ds.client.RegisterDriver(ctx, req)
// }
