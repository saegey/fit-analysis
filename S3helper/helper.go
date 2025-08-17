package S3helper

import (
	"bytes"
	"encoding/json"
	"fmt"

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
