package srvn

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

//go test -v -run=TestConsulDial
func TestConsulDial(t *testing.T) {
	endpoints := []string{"127.0.0.1:8500"}

	// var cfg Config
	cfg := Config{Provider: Consul, Endpoints: endpoints}
	fmt.Println(cfg)

	s, err := NewDiscover(cfg)
	if err != nil {
		panic(err)
	}

	conn, err := s.Dial("go_service_passport")
	if err != nil {
		panic(err)
	}

	fmt.Println(conn)
	Greeting(conn)
}

//go test -v -run=TestGRPC
func TestGRPC(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	Greeting(conn)
}

// Usage:
// cd $GOPATH/src/google.golang.org/grpc/examples/helloworld
// Startup a greeter_server of the grpc examples
func Greeting(conn *grpc.ClientConn) {
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := "world"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %s", r.Message)
}
