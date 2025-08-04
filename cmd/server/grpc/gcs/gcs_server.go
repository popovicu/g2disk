package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/popovicu/g2disk/proto"
)

type GCSServer struct {
	client      *storage.Client
	bucket      string
	chunkSize   uint64
	totalSize   uint64
	chunkPrefix string
}

func (s *GCSServer) getChunkName(chunkIdx uint64) string {
	return fmt.Sprintf("%s_%08d", s.chunkPrefix, chunkIdx)
}

func (s *GCSServer) readChunk(ctx context.Context, chunkIdx uint64) ([]byte, error) {
	objName := s.getChunkName(chunkIdx)
	slog.Info("Attempting to read chunk", "bucket", s.bucket, "object", objName, "chunkIdx", chunkIdx)

	obj := s.client.Bucket(s.bucket).Object(objName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			slog.Warn("Chunk doesn't exist in GCS, returning zeros",
				"bucket", s.bucket,
				"object", objName,
				"error", err)
			return make([]byte, s.chunkSize), nil
		}
		slog.Error("Failed to create reader",
			"bucket", s.bucket,
			"object", objName,
			"error", err)
		return nil, fmt.Errorf("failed to read chunk %d: %w", chunkIdx, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("Failed to read data from reader",
			"object", objName,
			"error", err)
		return nil, fmt.Errorf("failed to read chunk %d data: %w", chunkIdx, err)
	}

	// Log first few bytes to verify we got real data
	sample := ""
	if len(data) >= 16 {
		sample = fmt.Sprintf("%x", data[:16])
	}
	slog.Info("Successfully read chunk",
		"object", objName,
		"size", len(data),
		"first16bytes", sample)

	return data, nil
}

func (s *GCSServer) writeChunk(ctx context.Context, chunkIdx uint64, data []byte) error {
	objName := s.getChunkName(chunkIdx)
	obj := s.client.Bucket(s.bucket).Object(objName)

	writer := obj.NewWriter(ctx)
	writer.ChunkSize = 0 // This disables chunked/multipart uploads
	_, err := writer.Write(data)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write chunk %d: %w", chunkIdx, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer for chunk %d: %w", chunkIdx, err)
	}

	return nil
}

func (s *GCSServer) GetSize(ctx context.Context, req *proto.GetSizeRequest) (*proto.GetSizeResponse, error) {
	return &proto.GetSizeResponse{
		Size: s.totalSize,
	}, nil
}

func (s *GCSServer) Read(ctx context.Context, req *proto.ReadRequest) (*proto.ReadResponse, error) {
	offset := req.GetOffset()
	readSize := req.GetReadSize()

	slog.Info("Read request", "offset", offset, "size", readSize)

	if offset >= s.totalSize {
		return nil, fmt.Errorf("read offset %d out of bounds for disk size %d", offset, s.totalSize)
	}

	if offset+readSize > s.totalSize {
		return nil, fmt.Errorf("read request (offset %d, size %d) out of bounds for disk size %d",
			offset, readSize, s.totalSize)
	}

	result := make([]byte, readSize)
	bytesRead := uint64(0)

	for bytesRead < readSize {
		currentOffset := offset + bytesRead
		chunkIdx := currentOffset / s.chunkSize
		chunkOffset := currentOffset % s.chunkSize

		// Read the chunk
		chunkData, err := s.readChunk(ctx, chunkIdx)
		if err != nil {
			return nil, err
		}

		// Calculate how much to read from this chunk
		remainingInChunk := s.chunkSize - chunkOffset
		toRead := readSize - bytesRead
		if toRead > remainingInChunk {
			toRead = remainingInChunk
		}

		// Copy data
		copy(result[bytesRead:bytesRead+toRead], chunkData[chunkOffset:chunkOffset+toRead])
		bytesRead += toRead
	}

	return &proto.ReadResponse{
		Payload: result,
	}, nil
}

func (s *GCSServer) Write(ctx context.Context, req *proto.WriteRequest) (*proto.WriteResponse, error) {
	offset := req.GetOffset()
	payload := req.GetPayload()
	writeSize := uint64(len(payload))

	if offset >= s.totalSize {
		return nil, fmt.Errorf("write offset %d out of bounds for disk size %d", offset, s.totalSize)
	}

	if offset+writeSize > s.totalSize {
		return nil, fmt.Errorf("write request (offset %d, size %d) out of bounds for disk size %d",
			offset, writeSize, s.totalSize)
	}

	bytesWritten := uint64(0)

	for bytesWritten < writeSize {
		currentOffset := offset + bytesWritten
		chunkIdx := currentOffset / s.chunkSize
		chunkOffset := currentOffset % s.chunkSize

		// Read the existing chunk (or get zeros if it doesn't exist)
		chunkData, err := s.readChunk(ctx, chunkIdx)
		if err != nil {
			return nil, err
		}

		// Calculate how much to write to this chunk
		remainingInChunk := s.chunkSize - chunkOffset
		toWrite := writeSize - bytesWritten
		if toWrite > remainingInChunk {
			toWrite = remainingInChunk
		}

		// Modify the chunk
		copy(chunkData[chunkOffset:chunkOffset+toWrite], payload[bytesWritten:bytesWritten+toWrite])

		// Write the modified chunk back to GCS
		if err := s.writeChunk(ctx, chunkIdx, chunkData); err != nil {
			return nil, err
		}

		bytesWritten += toWrite
	}

	return &proto.WriteResponse{}, nil
}

func bucketExists(ctx context.Context, client *storage.Client, bucketName string) (bool, error) {
	_, err := client.Bucket(bucketName).Attrs(ctx)
	if err == storage.ErrBucketNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func createBucket(ctx context.Context, client *storage.Client, bucketName string) error {
	bucket := client.Bucket(bucketName)

	// Create bucket with minimal configuration
	// For emulator testing, we don't need to specify location or other attributes
	err := bucket.Create(ctx, "testing_project", nil)
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}

	slog.Info("Successfully created bucket", "bucket", bucketName)
	return nil
}

func ensureBucket(ctx context.Context, client *storage.Client, bucketName string, autoCreate bool) error {
	exists, err := bucketExists(ctx, client, bucketName)
	if err != nil {
		// For emulators, bucket.Attrs() might fail in unexpected ways
		// Try to list objects as a fallback check
		it := client.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: "dummy"})
		_, err := it.Next()
		if err == iterator.Done || err == nil {
			// Bucket exists (either empty or has objects)
			slog.Info("Bucket exists (verified via object listing)", "bucket", bucketName)
			return nil
		}
		if err != storage.ErrBucketNotExist {
			slog.Warn("Failed to check bucket existence via Attrs, assuming it needs creation",
				"bucket", bucketName, "error", err)
			exists = false
		} else {
			return fmt.Errorf("failed to check if bucket exists: %w", err)
		}
	}

	if exists {
		slog.Info("Bucket already exists", "bucket", bucketName)
		return nil
	}

	if !autoCreate {
		return fmt.Errorf("bucket %s does not exist and --auto-create-bucket flag is not set", bucketName)
	}

	slog.Info("Bucket does not exist, attempting to create it", "bucket", bucketName)
	return createBucket(ctx, client, bucketName)
}

func main() {
	var (
		bucket           = flag.String("bucket", "", "GCS bucket name")
		chunkSize        = flag.Uint64("chunk-size", 4*1024*1024, "Chunk size in bytes (default 4MB)")
		totalSize        = flag.Uint64("total-size", 100*1024*1024, "Total disk size in bytes (default 100MB)")
		chunkPrefix      = flag.String("chunk-prefix", "chunk", "Prefix for chunk object names")
		gcsEndpoint      = flag.String("gcs-endpoint", "", "Custom GCS endpoint (for testing)")
		port             = flag.String("port", "", "Port to listen on (overrides PORT env var)")
		autoCreateBucket = flag.Bool("auto-create-bucket", false, "Create bucket if it doesn't exist")
	)

	flag.Parse()

	if *bucket == "" {
		log.Fatal("--bucket flag is required")
	}

	if *totalSize%*chunkSize != 0 {
		log.Fatalf("Total size (%d) must be a multiple of chunk size (%d)", *totalSize, *chunkSize)
	}

	logger := slog.Default()

	// Determine port
	listenPort := *port
	if listenPort == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			listenPort = envPort
		} else {
			listenPort = "8080"
		}
	}

	ctx := context.Background()

	// Create GCS client
	var opts []option.ClientOption
	if *gcsEndpoint != "" {
		opts = append(opts, option.WithEndpoint(*gcsEndpoint), option.WithoutAuthentication())
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	// Ensure bucket exists
	if err := ensureBucket(ctx, client, *bucket, *autoCreateBucket); err != nil {
		log.Fatalf("Bucket initialization failed: %v", err)
	}

	server := &GCSServer{
		client:      client,
		bucket:      *bucket,
		chunkSize:   *chunkSize,
		totalSize:   *totalSize,
		chunkPrefix: *chunkPrefix,
	}

	// Setup gRPC server
	target := fmt.Sprintf("0.0.0.0:%s", listenPort)
	logger.Info("Starting GCS-backed disk server",
		"target", target,
		"bucket", *bucket,
		"chunkSize", *chunkSize,
		"totalSize", *totalSize,
		"chunks", *totalSize / *chunkSize,
		"chunkPrefix", *chunkPrefix)

	lis, err := net.Listen("tcp", target)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterG2DiskServiceServer(grpcServer, server)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
