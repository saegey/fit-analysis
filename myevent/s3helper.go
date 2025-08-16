package myevent

import (
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// getObjectMetadata retrieves the metadata of an object in an S3 bucket.
//
// Parameters:
// - svc: An instance of the S3 service client.
// - bucket: The name of the S3 bucket.
// - key: The key of the object in the S3 bucket.
//
// Returns:
// - *s3.HeadObjectOutput: The metadata of the object.
// - error: An error if the operation fails.
func GetObjectMetadata(svc *s3.S3, bucket string, key string) (*s3.HeadObjectOutput, error) {
	metaData, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	log.Printf("Metadata: %v", metaData)
	return metaData, err
}

func GetObjectBody(svc *s3.S3, bucket string, key string) ([]byte, error) {
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %v", err)
	}

	return body, nil
}
