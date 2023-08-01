package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	pb "live/pb"

	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedStreamingServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":8081")
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
	}
	return nil
}

// func (s *server) GetLiveStream(stream pb.StreamingService_GetLiveStreamServer) error {
// 	absFilePath, err := filepath.Abs("temp_video.mp4")
// 	if err != nil {
// 		log.Fatalf("failed to get absolute file path: %v", err)
// 	}
// 	videoURL := "http://localhost:8080/video/streetacademy.mp4"
// 	if err := downloadVideo(videoURL, absFilePath); err != nil {
// 		return fmt.Errorf("failed to download video: %v", err)
// 	}
// 	defer os.Remove(absFilePath)
// 	return nil
// }

func fetchVideoStream(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	return resp.Body, nil
}

func downloadVideo(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download video: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save video file: %v", err)
	}

	return nil
}
