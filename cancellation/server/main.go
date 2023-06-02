package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	pb "github.com/leemingeer/grpc-simple/lib/helloworld/helloworld"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "port number")

type server struct {
	pb.UnimplementedHelloServer
	client pb.HelloClient
	cc     *grpc.ClientConn
}

func (s *server) UnaryHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	message := req.GetName()
	if strings.HasPrefix(message, "[propagate me]") {
		time.Sleep(800 * time.Millisecond)
		message = strings.TrimPrefix(message, "[propagate me]")
		return s.client.UnaryHello(ctx, &pb.HelloRequest{Name: message})
	}

	if message == "delay" {
		time.Sleep(2 * time.Second)
	}

	return &pb.HelloReply{Message: req.GetName()}, nil
}

func (s *server) BidirectionalStreamingHello(stream pb.Hello_BidirectionalStreamingHelloServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			fmt.Printf("server: error receiving from stream: %v\n", err)
			if err == io.EOF {
				return nil
			}
			return err
		}
		fmt.Printf("echoing message %q\n", in.GetName())
		stream.Send(&pb.HelloReply{Message: in.GetName()})
	}
}

func (s *server) Close() {
	s.cc.Close()
}

func main() {
	flag.Parse()

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServer(grpcServer, &server{})
	fmt.Printf("server listening at port %v\n", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
