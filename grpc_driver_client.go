package main

import (
	"fmt"
	"log"

	pb "github.com/ride4Low/contracts/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DriverServiceClient wraps the gRPC client for DriverService
type DriverServiceClient struct {
	conn   *grpc.ClientConn
	client pb.DriverServiceClient
}

// NewDriverServiceClient creates a new DriverServiceClient
func NewDriverServiceClient(address string) (*DriverServiceClient, error) {
	// Establish connection to the driver service
	log.Println("Connecting to driver service at:", address)

	// conn, err := grpc.Dial(address,
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// 	grpc.WithBlock(),
	// 	grpc.WithTimeout(5*time.Second),
	// )
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to driver service: %w", err)
	}

	client := pb.NewDriverServiceClient(conn)

	log.Println("connected")

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
