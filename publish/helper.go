package publish

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
)

// publishMessage publishes a message using the IoT Data Plane client.
// It takes an IoT Data Plane client as input and a postId as a string pointer.
// The function returns an error if there was a problem publishing the message.
func PublishMessage(iotClient *iotdataplane.IoTDataPlane, postId *string, message string) error {

	var payload = fmt.Sprintf(`{"phase": "%s"}`, message)
	input := &iotdataplane.PublishInput{
		Topic:   aws.String(fmt.Sprintf("post-%s", *postId)),
		Payload: []byte(payload),
	}

	_, err := iotClient.Publish(input)
	return err
}
