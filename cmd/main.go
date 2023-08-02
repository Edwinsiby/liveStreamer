package main

import (
	"fmt"
	"log"
	"net"

	pb "live/pb"
	service "live/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	myService := &service.Server{}
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStreamingServiceServer(s, myService)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("live stream server is running in 8081")
}
