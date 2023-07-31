package main

import (
	"log"
	"net"
	"time"

	pb "live/pb"

	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedStreamingServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStreamingServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) GetLiveStream(stream pb.StreamingService_GetLiveStreamServer) error {

	videoCapture, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("failed to open camera: %v", err)
	}
	frame := gocv.NewMat()
	defer frame.Close()

	for {
		if ok := videoCapture.Read(&frame); !ok {
			break
		}

		data, err := gocv.IMEncode(".jpg", frame)
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.StreamingResponse{Data: data.GetBytes()}); err != nil {
			return err
		}

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}
