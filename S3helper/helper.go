package S3helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lambda/myevent"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type UploadParams struct {
	Powers []float64
	Bucket string
	PostID string
}

type MergedDataItem struct {
	Power     uint16  `json:"p"`
	Distance  float64 `json:"d"`
	Time      float64 `json:"t"`
	Elevation float32 `json:"e"`
	HeartRate uint8   `json:"h"`
	Grade     float64 `json:"g"`
}

type UploadToS3Input struct {
	Coordinates [][]float64
	Elevation   []MergedDataItem
	Bucket      string
	IdentityID  string
	S3Key       string
}

func UploadToS3(input UploadToS3Input) error {
	fmt.Println("Uploading object to S3")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	svc := s3.New(sess)

	fmt.Println("S3 key: ", input.S3Key)

	data := map[string]interface{}{
		"coordinates": input.Coordinates,
		"elevation":   input.Elevation,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %v", err)
	}
	var s3filename = fmt.Sprintf("private/%s/%s", input.IdentityID, input.S3Key)
	fmt.Println("S3 filename: ", s3filename)
	fmt.Println("Bucket: ", input.Bucket)

	res, err := svc.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(jsonData)),
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(s3filename),
	})
	fmt.Println("Response: ", res)

	if err != nil {
		fmt.Printf("failed to upload object to S3: %v\n", err)
	} else {
		fmt.Println("Response: ", res)
		fmt.Println("Object uploaded successfully")
	}

	return nil
}

type MetaData struct {
	Key        string
	Bucket     string
	PostId     string
	IdentityId string
}

func ExtractMetaData(record myevent.Record, svc *s3.S3) (MetaData, error) {
	sanitizedKey, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		return MetaData{}, fmt.Errorf("error unescaping string: %v", err)
	}

	var bucket = record.S3.Bucket.Name
	log.Printf("Processing S3 object from bucket %s with key %s", bucket, sanitizedKey)

	// Get the metadata of the S3 object
	metaData, err := myevent.GetObjectMetadata(svc, bucket, sanitizedKey)
	if err != nil {
		return MetaData{}, fmt.Errorf("failed to get metadata: %v", err)
	}
	var postId = metaData.Metadata["Postid"]
	if postId == nil {
		return MetaData{}, fmt.Errorf("missing post id: %v", err)
	}

	var identityId = metaData.Metadata["Identityid"]
	if identityId == nil {
		return MetaData{}, fmt.Errorf("missing identity id: %v", err)
	}
	return MetaData{Bucket: bucket, PostId: *postId, IdentityId: *identityId, Key: sanitizedKey}, nil
}
