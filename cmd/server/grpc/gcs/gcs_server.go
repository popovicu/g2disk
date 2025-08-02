package main

// TODO: this is all still just a ramdisk server, should be actual GCS

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/popovicu/g2disk/proto"
)

type RamdiskServer struct {
	disk []byte
}

func (s *RamdiskServer) GetSize(ctx context.Context, req *proto.GetSizeRequest) (*proto.GetSizeResponse, error) {
	return &proto.GetSizeResponse{
		Size: uint64(len(s.disk)),
	}, nil
}

func (s *RamdiskServer) Read(ctx context.Context, req *proto.ReadRequest) (*proto.ReadResponse, error) {
	offset := req.GetOffset()
	readSize := req.GetReadSize()

	diskSize := uint64(len(s.disk))

	if offset >= diskSize {
		return nil, fmt.Errorf("read offset %d out of bounds for disk size %d", offset, diskSize)
	}

	if offset+readSize > diskSize {
		return nil, fmt.Errorf("read request (offset %d, size %d) out of bounds for disk size %d", offset, readSize, diskSize)
	}

	payload := s.disk[offset : offset+readSize]

	return &proto.ReadResponse{
		Payload: payload,
	}, nil
}

func (s *RamdiskServer) Write(ctx context.Context, req *proto.WriteRequest) (*proto.WriteResponse, error) {
	offset := req.GetOffset()
	payload := req.GetPayload()
	writeSize := uint64(len(payload))

	diskSize := uint64(len(s.disk))

	if offset >= diskSize {
		return nil, fmt.Errorf("write offset %d out of bounds for disk size %d", offset, diskSize)
	}

	if offset+writeSize > diskSize {
		return nil, fmt.Errorf("write request (offset %d, size %d) out of bounds for disk size %d", offset, writeSize, diskSize)
	}

	copy(s.disk[offset:offset+writeSize], payload)

	return &proto.WriteResponse{}, nil
}

func main() {
	flag.Parse()
	logger := slog.Default()

	port, portEnvExists := os.LookupEnv("PORT")

	if !portEnvExists {
		port = "8080"
	}

	target := fmt.Sprintf("0.0.0.0:%s", port)
	logger.Info("Listening on a TCP target", "target", target)

	lis, err := net.Listen("tcp", target)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &RamdiskServer{
		disk: make([]byte, 5*1024*1024),
	}
	proto.RegisterG2DiskServiceServer(grpcServer, server)

	logger.Info("Starting the API sever", "target", target)
	grpcServer.Serve(lis)
}
