package main

import (
	"context"
	"encoding/json"

	"github.com/ride4Low/contracts/events"
)

type MessagePublisher interface {
	PublishMessage(ctx context.Context, routingKey string, message events.AmqpMessage) error
}

type EventPublisher interface {
	PublishWithData(ctx context.Context, routingKey string, userID string, data json.RawMessage) error
	PublishStripeCreateSession(ctx context.Context, userID string, data json.RawMessage) error
}

type AmqpPublisher struct {
	publisher MessagePublisher
}

func NewAmqpPublisher(publisher MessagePublisher) *AmqpPublisher {
	return &AmqpPublisher{publisher: publisher}
}

func (p *AmqpPublisher) PublishWithData(ctx context.Context, routingKey string, userID string, data json.RawMessage) error {
	return p.publisher.PublishMessage(ctx, routingKey, events.AmqpMessage{
		OwnerID: userID,
		Data:    data,
	})
}

func (p *AmqpPublisher) PublishStripeCreateSession(ctx context.Context, userID string, data json.RawMessage) error {
	return p.publisher.PublishMessage(ctx, events.PaymentCmdCreateSession, events.AmqpMessage{
		OwnerID: userID,
		Data:    data,
	})
}
