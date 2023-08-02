package service

import (
	"fmt"
	"io"
	pb "live/pb"
	"log"
	"net/http"
	"os"
	"sync"

	"gocv.io/x/gocv"
)

type Server struct {
	pb.UnimplementedStreamingServiceServer
}

func (s *Server) GetLiveStream(stream pb.StreamingService_GetLiveStreamServer) error {
	ctx := stream.Context()

	videoCapture, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("failed to open camera: %v", err)
	}

	frame := gocv.NewMat()

	stopChan := make(chan struct{})
	var once sync.Once

	go func() {
		defer func() {
			videoCapture.Close()
			frame.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				once.Do(func() {
					close(stopChan)
				})
				return
			default:
				if ok := videoCapture.Read(&frame); !ok {
					return
				}
				data, err := gocv.IMEncode(".jpg", frame)
				if err != nil {
					log.Printf("failed to encode frame: %v", err)
					return
				}
				if err := stream.Send(&pb.StreamingResponse{Data: data.GetBytes()}); err != nil {
					log.Printf("failed to send frame to client: %v", err)
					return
				}
			}
		}
	}()

	<-stopChan

	return nil
}

func FetchVideoStream(url string) (io.ReadCloser, error) {
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

func DownloadVideo(url, filePath string) error {
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
