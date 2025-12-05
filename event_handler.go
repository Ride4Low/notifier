package main

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ride4Low/contracts/events"
)

type EventHandler struct {
	cm *ConnectionManager
}

func NewEventHandler(cm *ConnectionManager) *EventHandler {
	return &EventHandler{cm: cm}
}

func (h *EventHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	var message events.AmqpMessage

	if msg.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	if err := sonic.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	switch msg.RoutingKey {
	// rider ws consume this.
	case events.TripEventNoDriversFound:
		return h.handleNoDriversFound(ctx, message)
	default:
		return fmt.Errorf("unknown routing key: %s", msg.RoutingKey)
	}
}

func (h *EventHandler) handleNoDriversFound(ctx context.Context, message events.AmqpMessage) error {
	if message.OwnerID == "" {
		return fmt.Errorf("owner ID is empty")
	}

	if message.Data == nil {
		return fmt.Errorf("data is nil")
	}

	var payload any

	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	wsMessage := WSMessage{
		Type: events.TripEventNoDriversFound,
		Data: payload,
	}

	err := h.cm.SendMessage(message.OwnerID, wsMessage)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// func (h *EventHandler) handleTripRequest(ctx context.Context, message events.AmqpMessage) error {

// 	if message.OwnerID == "" {
// 		return fmt.Errorf("owner ID is empty")
// 	}

// 	if message.Data == nil {
// 		return fmt.Errorf("data is nil")
// 	}

// 	var trip events.TripEventData

// 	if err := sonic.Unmarshal(message.Data, &trip); err != nil {
// 		return fmt.Errorf("failed to unmarshal message: %v", err)
// 	}

// 	wsMessage := WSMessage{
// 		Type: "trip_request",
// 		Data: trip,
// 	}

// 	return nil
// }
