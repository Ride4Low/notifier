package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ride4Low/contracts/pkg/otel"
	pb "github.com/ride4Low/contracts/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// DriverServiceClient wraps the gRPC client for DriverService
type DriverServiceClient struct {
	conn   *grpc.ClientConn
	client pb.DriverServiceClient
}

// func validateTarget(addr string, timeout time.Duration) error {
// 	conn, err := net.DialTimeout("tcp", addr, timeout)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to driver service: %w", err)
// 	}
// 	conn.Close()
// 	return nil
// }

// NewDriverServiceClient creates a new DriverServiceClient
func NewDriverServiceClient(address string) (*DriverServiceClient, error) {
	// Establish connection to the driver service
	log.Println("Connecting to driver service at:", address)

	// if err := validateTarget(address, 5*time.Second); err != nil {
	// 	log.Println("Failed to connect to driver service:", err)
	// 	return nil, err
	// }

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	dialOptions = append(dialOptions, otel.ClientOptions()...)

	conn, err := grpc.NewClient(address, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to driver service: %w", err)
	}

	if err := waitForReady(conn); err != nil {
		return nil, fmt.Errorf("failed to connect to driver service: %w", err)
	}

	log.Println("connected")

	client := pb.NewDriverServiceClient(conn)

	return &DriverServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *DriverServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// func (c *DriverServiceClient) GetState() connectivity.State {
// 	return c.conn.GetState()
// }

func waitForReady(conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn.Connect()

	for {
		state := conn.GetState()
		log.Println("Current state:", state)

		if state == connectivity.Ready {
			log.Println("Connection is READY!")
			return nil
		}

		// Block until state changes OR context times out
		if !conn.WaitForStateChange(ctx, state) {
			log.Println("Timeout waiting for state change")
			return fmt.Errorf("timeout waiting for state change")
		}
	}
}
