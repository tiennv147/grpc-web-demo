package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/crypto-game-portal/playground/grpc-web-demo/baucua/pb"
	"google.golang.org/grpc"
)

const (
	listenAddress = "0.0.0.0:9090"
)

type greeterServer struct {
}

func newServer() *greeterServer {
	return &greeterServer{}
}

func (*greeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello! %s", req.Name),
	}, nil
}
func (*greeterServer) SayRepeatHello(req *pb.RepeatHelloRequest, stream pb.Greeter_SayRepeatHelloServer) error {

	for i := 0; i < int(req.Count); i++ {
		stream.Send(&pb.HelloReply{
			Message: fmt.Sprintf("Hello! %s %d from Go", req.Name, i+1),
		})
		time.Sleep(1 * time.Second)
	}
	return nil
}

func handleSignals(chnl <-chan os.Signal, svr *grpc.Server) {
	for sig := range chnl {
		go func(sig os.Signal) {
			if sig == syscall.SIGTERM || sig == syscall.SIGKILL {
				svr.GracefulStop()
			}
		}(sig)
	}
}

func main() {
	chanSignal := make(chan os.Signal, 1)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		err = fmt.Errorf("failed to listen on the address %s: %v", listenAddress, err)
	}

	gRPCServer := grpc.NewServer()
	signal.Notify(chanSignal)
	go handleSignals(chanSignal, gRPCServer)

	pb.RegisterGreeterServer(gRPCServer, newServer())

	fmt.Println("start gRPC Serer on ", listenAddress)
	if err = gRPCServer.Serve(listener); err != nil {
		log.Fatalf("failed to launch gRPCServer on address: %v", err)
	}
}
