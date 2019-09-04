package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

const (
	defaultName = "world"
)

func main() {
	lis, err := net.Listen("tcp", ":50050")
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	service.RegisterChannelzServiceToServer(s)
	go s.Serve(lis)
	defer s.Stop()

	// Initialize manual resolver and Dial.
	r, rcleanup := manual.GenerateAndRegisterManualResolver()
	r.InitialState(resolver.State{
		Addresses: []resolver.Address{
			{Addr: ":10001"},
			{Addr: ":10002"},
			{Addr: ":10003"},
		},
	})
	defer rcleanup()
	// Set up a connection to the server.
	conn, err := grpc.Dial(r.Scheme()+":///test.server", grpc.WithInsecure(), grpc.WithBalancerName("round_robin"))
	if err != nil {
		log.Fatal("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	// Make 100 SayHello RPCs.
	for i := 0; i < 100; i++ {
		// Setting a 150ms timeout on the RPC.
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("could not greet: %v", err)
		} else {
			log.Printf("Greeting: %s", r.Message)
		}
	}

	// Wait for CTRL+C to exit.
	// Unless you exit the program with CTRL+C, channelz data will be available for querying.
	// Users can take time to examine and learn about the info provided by channelz.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block until a signal is received.
	<-ch
}
