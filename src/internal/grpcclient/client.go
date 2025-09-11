package grpcclient

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/Milad-Abooali/4in-cs2skin-g1/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client pb.DataServiceClient

// Connect establishes a gRPC connection to the Core service
func Connect(address string) {
	log.Println("Connecting to Core gRPC:", address)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC Core: %v", err)
	}

	// Make sure the connection is ready before continuing
	state := conn.GetState()
	if !conn.WaitForStateChange(ctx, state) {
		log.Fatalf("Connection not ready before timeout")
	}

	client = pb.NewDataServiceClient(conn)
	log.Println("✅ Connected to gRPC Core:", address)
}

// SendQuery sends a raw SQL query to the Core with access token
func SendQuery(query string) (*pb.QueryResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := os.Getenv("CORE_GRPC_TOKEN")

	return client.Query(ctx, &pb.QueryRequest{
		Token: token,
		Query: query,
	})
}

// TestConnection performs a simple test query to verify gRPC connectivity
func TestConnection() {
	resp, err := SendQuery("SELECT version();")
	if err != nil {
		log.Printf("❌ gRPC test failed: %v", err)
		return
	}
	log.Printf("✅ gRPC test successful: %v", resp)
}
