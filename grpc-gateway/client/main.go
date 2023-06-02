package main

import (
	"flag"
	"log"

	pb "github.com/leemingeer/grpc-simple/lib/helloworld/helloworld"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", "0.0.0.0:50051", "the address to connect to")

func main() {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)
	r, err := c.UnaryHello(context.Background(), &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
