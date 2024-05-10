package main

import (
	"fmt"
	"log"
	"net"
	"udev-broker-golang/broker"

	pb "udev-broker-golang/grpc"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
)

const LISTEN_PORT = 50050

// Assume mjpeg input
func main() {
	// init log
	Logger, _ := broker.InitLogger()
	// init broker
	broker.GetDevPath()
	broker := broker.InitBroker()
	Logger.Debugf("broker: %#v", broker)

	// init grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", LISTEN_PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCameraServer(grpcServer, broker)
	grpcServer.Serve(lis)
}
