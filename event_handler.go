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
	// driver ws consume this.
	case events.DriverCmdTripRequest:
		return h.handleTripRequest(ctx, message)
	default:
		return fmt.Errorf("unknown routing key: %s", msg.RoutingKey)
	}
}

func (h *EventHandler) handleNoDriversFound(_ context.Context, message events.AmqpMessage) error {
	return h.sendWSMessage(message, events.TripEventNoDriversFound)
}

func (h *EventHandler) handleTripRequest(_ context.Context, message events.AmqpMessage) error {
	return h.sendWSMessage(message, events.DriverCmdTripRequest)
}

// sendWSMessage validates the message, unmarshals the payload, and sends it via WebSocket.
func (h *EventHandler) sendWSMessage(message events.AmqpMessage, eventType string) error {
	if message.OwnerID == "" {
		return fmt.Errorf("owner ID is empty")
	}

	if message.Data == nil {
		return fmt.Errorf("data is nil")
	}

	var payload any
	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	wsMessage := WSMessage{
		Type: eventType,
		Data: payload,
	}

	if err := h.cm.SendMessage(message.OwnerID, wsMessage); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
