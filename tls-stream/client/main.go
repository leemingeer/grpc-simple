package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "github.com/leemingeer/grpc-simple/lib/helloworld/helloworld"
	"github.com/leemingeer/grpc-simple/utils/authentication"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	address = ":50051"
)

func unary(client pb.HelloClient) {
	r, err := client.UnaryHello(context.Background(), &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

}

// unaryInterceptor 一个简单的 unary interceptor 示例。
func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// pre-processing
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...) // invoking RPC method
	// post-processing
	end := time.Now()
	log.Printf("RPC: %s, req:%v start time: %s, end time: %s, err: %v", method, req, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
	return err
}

func serverStream(client pb.HelloClient) {
	// 2.调用获取stream
	stream, err := client.ServerStreamingHello(context.Background(), &pb.HelloRequest{Name: "Hello server stream"})
	if err != nil {
		log.Fatalf("could not echo: %v", err)
	}

	// 3. for循环获取服务端推送的消息
	for {
		// 通过 Recv() 不断获取服务端send()推送的消息
		resp, err := stream.Recv()
		// 4. err==io.EOF则表示服务端关闭stream了 退出
		if err == io.EOF {
			log.Println("server closed")
			break
		}
		if err != nil {
			log.Printf("Recv error:%v", err)
			continue
		}
		log.Printf("Recv data:%v", resp.GetMessage())
	}
}

func clientStream(client pb.HelloClient) {
	// 2.获取 stream 并通过 Send 方法不断推送数据到服务端
	stream, err := client.ClientStreamingHello(context.Background())
	if err != nil {
		log.Fatalf("Sum() error: %v", err)
	}
	for i := 0; i < 5; i++ {
		err := stream.Send(&pb.HelloRequest{Name: "Hello client stream: " + strconv.Itoa(i)})
		if err != nil {
			log.Printf("send error: %v", err)
			continue
		}
	}

	// 3. 发送完成后通过stream.CloseAndRecv() 关闭steam并接收服务端返回结果
	// (服务端则根据err==io.EOF来判断client是否关闭stream)
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("CloseAndRecv() error: %v", err)
	}
	log.Printf("sum: %v", resp.GetMessage())

}

func bidirectionalStream(client pb.HelloClient) {
	var wg sync.WaitGroup
	// 2. 调用方法获取stream
	stream, err := client.BidirectionalStreamingHello(context.Background())
	if err != nil {
		panic(err)
	}
	// 3.开两个goroutine 分别用于Recv()和Send()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("Server Closed")
				break
			}
			if err != nil {
				continue
			}
			fmt.Printf("Recv Bidirectional Data:%v \n", req.GetMessage())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < 2; i++ {
			err := stream.Send(&pb.HelloRequest{Name: "hello bi sent by client:" + strconv.Itoa(i)})
			if err != nil {
				log.Printf("send error:%v\n", err)
			}
			time.Sleep(time.Second)
		}
		// 4. 发送完毕关闭stream
		err := stream.CloseSend()
		if err != nil {
			log.Printf("Send error:%v\n", err)
			return
		}
	}()
	wg.Wait()
}

// wrappedStream  用于包装 grpc.ClientStream 结构体并拦截其对应的方法。
type wrappedStream struct {
	grpc.ClientStream
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

// streamInterceptor 一个简单的 stream interceptor 示例。
func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	// 返回的是自定义的封装过的 stream
	return newWrappedStream(s), nil
}

func main() {

	// 加载客户端证书
	certificate, err := tls.LoadX509KeyPair("../scripts/x509/client.crt", "../scripts/x509/client.key")
	if err != nil {
		log.Fatal(err)
	}
	// 构建CertPool以校验服务端证书有效性
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../scripts/x509/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   "www.ming.com", // NOTE: this is required!
		RootCAs:      certPool,
	})
	//构建一个 PerRPCCredentials。
	myAuth := authentication.NewMyAuth()
	conn, err := grpc.Dial(address, grpc.WithPerRPCCredentials(myAuth), grpc.WithTransportCredentials(creds), grpc.WithUnaryInterceptor(unaryInterceptor), grpc.WithStreamInterceptor(streamInterceptor), grpc.WithBlock())
	//conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)

	unary(c)

	//serverStream(c)
	//
	//clientStream(c)
	//
	//bidirectionalStream(c)
}
