package file

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	Writer(ctx context.Context, filePath string) (io.WriteCloser, error)
	Reader(ctx context.Context, filePath string) (io.ReadCloser, error)
}

func NewClient(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	switch downloadConfig.Mode {
	case configs.DownloadModeLocal:
		return NewLocalClient(downloadConfig, logger)
	case configs.DownloadModelS3:
		return NewS3Client(downloadConfig, logger)
	default:
		return nil, fmt.Errorf("unsupported download mode: %s", downloadConfig.Mode)
	}
}

// bufferedFileReader is used in LocalClient
type bufferedFileReader struct {
	file           *os.File
	bufferedReader io.Reader
}

func newBufferedFileReader(
	file *os.File,
) io.ReadCloser {
	return &bufferedFileReader{
		file:           file,
		bufferedReader: bufio.NewReader(file),
	}
}

func (b bufferedFileReader) Read(p []byte) (int, error) {
	return b.bufferedReader.Read(p)
}

func (b bufferedFileReader) Close() error {
	return b.file.Close()
}

type localClient struct {
	downloadDirectory string
	logger            *zap.Logger
}

func NewLocalClient(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	if err := os.MkdirAll(downloadConfig.DownloadDirectory, os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("failed to create download directory: %w", err)
		}
	}

	return &localClient{
		downloadDirectory: downloadConfig.DownloadDirectory,
		logger:            logger,
	}, nil
}

func (l localClient) Writer(ctx context.Context, filePath string) (io.WriteCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger)

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Create(absolutePath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create file")
		return nil, status.Error(codes.Internal, "failed to create file")
	}

	return file, nil
}

func (l localClient) Reader(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger)

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Open(absolutePath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open file")
		return nil, status.Error(codes.Internal, "failed to open file")
	}

	return newBufferedFileReader(file), nil
}

// s3ClientReadWriteCloser is used in S3Client
type s3ClientReadWriteCloser struct {
	writtenData []byte
	isClosed    bool
}

func newS3ClientReadWriteCloser(
	ctx context.Context,
	minioClient *minio.Client,
	logger *zap.Logger,
	bucketName,
	objectName string,
) io.ReadWriteCloser {
	logger = utils.LoggerWithContext(ctx, logger)
	readWriteCloser := &s3ClientReadWriteCloser{
		writtenData: make([]byte, 0),
		isClosed:    false,
	}

	go func() {
		if _, err := minioClient.PutObject(ctx, bucketName, objectName, readWriteCloser, -1, minio.PutObjectOptions{}); err != nil {
			logger.With(zap.Error(err)).Error("failed to put object")
		}
	}()

	return readWriteCloser
}

func (s *s3ClientReadWriteCloser) Read(p []byte) (int, error) {
	if len(s.writtenData) > 0 {
		// Copy writtenData to p, so s3 object can be get from p to process PutObject
		writtenLength := copy(p, s.writtenData)
		// reset writtenData to empty string
		s.writtenData = s.writtenData[writtenLength:]
		return writtenLength, nil
	}

	if s.isClosed {
		return 0, io.EOF
	}

	return 0, nil
}

func (s *s3ClientReadWriteCloser) Close() error {
	s.isClosed = true
	return nil
}

func (s *s3ClientReadWriteCloser) Write(p []byte) (int, error) {
	s.writtenData = append(s.writtenData, p...)
	return len(p), nil
}

type s3Client struct {
	minioClient *minio.Client
	bucket      string
	logger      *zap.Logger
}

func NewS3Client(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	minioClient, err := minio.New(downloadConfig.Address, &minio.Options{
		Creds:  credentials.NewStaticV4(downloadConfig.Username, downloadConfig.Password, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new minion client: %w", err)
	}

	if err := initBucket(context.Background(), minioClient, downloadConfig.Bucket); err != nil {
		return nil, fmt.Errorf("failed to init bucket: %w", err)
	}

	return &s3Client{
		minioClient: minioClient,
		bucket:      downloadConfig.Bucket,
		logger:      logger,
	}, nil
}

func initBucket(ctx context.Context, minioClient *minio.Client, bucketName string) error {
	ok, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !ok {
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (s s3Client) Reader(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, s.logger)

	object, err := s.minioClient.GetObject(ctx, s.bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get s3 object")
		return nil, status.Error(codes.Internal, "failed to get s3 object")
	}

	return object, nil
}

func (s s3Client) Writer(ctx context.Context, filePath string) (io.WriteCloser, error) {
	return newS3ClientReadWriteCloser(ctx, s.minioClient, s.logger, s.bucket, filePath), nil
}
