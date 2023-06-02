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

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/leemingeer/grpc-simple/lib/helloworld/helloworld"
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
		req, err := stream.Recv()
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "request message not received")
		}
		if err != nil {
			return err
		}

		message := req.GetName()
		if strings.HasPrefix(message, "[propagate me]") {
			time.Sleep(800 * time.Millisecond)
			message = strings.TrimPrefix(message, "[propagate me]")
			res, err := s.client.UnaryHello(stream.Context(), &pb.HelloRequest{Name: message})
			if err != nil {
				return err
			}
			stream.Send(res)
		}

		if message == "delay" {
			time.Sleep(2 * time.Second)
		}
		stream.Send(&pb.HelloReply{Message: message})
	}
}

func (s *server) Close() {
	s.cc.Close()
}

func newHelloServer() *server {
	target := fmt.Sprintf("localhost:%v", *port)
	cc, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return &server{client: pb.NewHelloClient(cc), cc: cc}
}

func main() {
	flag.Parse()

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	echoServer := newHelloServer()
	defer echoServer.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServer(grpcServer, echoServer)
	fmt.Printf("server listening at port %v\n", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
