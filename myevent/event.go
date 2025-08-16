package myevent

import (
	"context"
	"encoding/json"
	"fmt"
)

// Define the structs to match the JSON structure
type Event struct {
	Records []Record `json:"Records"`
}

type Record struct {
	EventVersion      string            `json:"eventVersion"`
	EventSource       string            `json:"eventSource"`
	AWSRegion         string            `json:"awsRegion"`
	EventTime         string            `json:"eventTime"`
	EventName         string            `json:"eventName"`
	UserIdentity      UserIdentity      `json:"userIdentity"`
	RequestParameters RequestParameters `json:"requestParameters"`
	ResponseElements  ResponseElements  `json:"responseElements"`
	S3                S3                `json:"s3"`
}

type UserIdentity struct {
	PrincipalID string `json:"principalId"`
}

type RequestParameters struct {
	SourceIPAddress string `json:"sourceIPAddress"`
}

type ResponseElements struct {
	XAmzRequestID string `json:"x-amz-request-id"`
	XAmzID2       string `json:"x-amz-id-2"`
}

type S3 struct {
	S3SchemaVersion string `json:"s3SchemaVersion"`
	ConfigurationID string `json:"configurationId"`
	Bucket          Bucket `json:"bucket"`
	Object          Object `json:"object"`
}

type Bucket struct {
	Name          string        `json:"name"`
	OwnerIdentity OwnerIdentity `json:"ownerIdentity"`
	ARN           string        `json:"arn"`
}

type OwnerIdentity struct {
	PrincipalID string `json:"principalId"`
}

type Object struct {
	Key       string `json:"key"`
	Size      int    `json:"size"`
	ETag      string `json:"eTag"`
	Sequencer string `json:"sequencer"`
}

// Sample handler function
func HandleRequest(ctx context.Context, event Event) (string, error) {
	// Example processing code
	fmt.Printf("Received event: %+v\n", event)
	return "Processed", nil
}

// Sample main function to test locally
func TestLocal() {
	jsonData := `{
        "Records": [
            {
                "eventVersion": "2.1",
                "eventSource": "aws:s3",
                "awsRegion": "us-east-1",
                "eventTime": "2024-08-08T18:21:12.554Z",
                "eventName": "ObjectCreated:Put",
                "userIdentity": {
                    "principalId": "AWS:AROAVN3XYARALPHSY2QQU:CognitoIdentityCredentials"
                },
                "requestParameters": {
                    "sourceIPAddress": "98.59.137.58"
                },
                "responseElements": {
                    "x-amz-request-id": "AVK5AAFTB69G8359",
                    "x-amz-id-2": "coiHYZTAMhzldJn7bjZY3S1q83nc3SZjC9TUDNsKiaFjLzIoh3XkPHyJcS7Kv61KRer+65LQLhWst+WdLtIjOsSLEa2bxJXntyWW5wdWrzU="
                },
                "s3": {
                    "s3SchemaVersion": "1.0",
                    "configurationId": "YmU5NWRlMTItZDVmZC00YTk1LWIxN2UtYTUxYWVmMmM5OTJj",
                    "bucket": {
                        "name": "nextjsblog619f9308251149be95e1a59d89c85e5e233823-dev",
                        "ownerIdentity": {
                            "principalId": "A16SA0H5NFM8TJ"
                        },
                        "arn": "arn:aws:s3:::nextjsblog619f9308251149be95e1a59d89c85e5e233823-dev"
                    },
                    "object": {
                        "key": "private/us-east-1%3Ad0793677-36a8-4f9d-96aa-af199ce7dd0f/uploads/RCC_Tahuya.fit",
                        "size": 1130542,
                        "eTag": "a62abf84c8dfd59a27ecb6092d3cd9e1",
                        "sequencer": "0066B50C983710C092"
                    }
                }
            }
        ]
    }`

	var event Event
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	response, err := HandleRequest(context.Background(), event)
	if err != nil {
		fmt.Println("Error handling request:", err)
		return
	}

	fmt.Println("Response:", response)
}
