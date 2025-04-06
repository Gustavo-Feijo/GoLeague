package logger

import (
	"context"
	"fmt"
	appConfig "goleague/pkg/config"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Logger that we will use to save our logs.
type NewLogger struct {
	mu       sync.Mutex
	logFile  *os.File
	filePath string
}

// Create the log instance with a temporary file.
func CreateLogger() (*NewLogger, error) {
	f, err := os.CreateTemp("", "log-*.log")
	if err != nil {
		return nil, err
	}

	return &NewLogger{
		logFile:  f,
		filePath: f.Name(),
	}, nil
}

// Log a simple info.
func (l *NewLogger) Infof(format string, args ...interface{}) {
	l.write("[INFO]", format, args...)
}

// Log a error.
func (l *NewLogger) Errorf(format string, args ...interface{}) {
	l.write("[ERROR]", format, args...)
}

// Write a empty line.
func (l *NewLogger) EmptyLine() {
	l.logFile.WriteString("\n")
}

// Write something to the logger.
func (l *NewLogger) write(infoType string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%-8s %s %s\n", infoType, timestamp, fmt.Sprintf(format, args...))

	l.logFile.WriteString(line)
}

// Clean the file contents.
func (l *NewLogger) CleanFile() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logFile.Truncate(0)

	l.logFile.Seek(0, 0)
}

// Upload the log to a s3 bucket.
func (l *NewLogger) UploadToS3Bucket(objectKey string) error {
	if _, err := l.logFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind file: %v", err)
	}

	// Get the config.
	cfg := aws.Config{
		Region: appConfig.Bucket.Region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				appConfig.Bucket.AccessKey,
				appConfig.Bucket.AccessSecret,
				"",
			),
		),
	}

	// Create the client.
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(appConfig.Bucket.Endpoint)
	})

	// Run the put.
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(appConfig.Bucket.LogBucket),
		Key:    aws.String(objectKey),
		Body:   l.logFile,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s to S3 bucket: %v", objectKey, err)
	}

	// Clean the file after sending.
	l.CleanFile()

	return nil
}
