package fitHelper

import (
	"bytes"
	"fmt"

	"github.com/tormoder/fit"
)

func DecodeFITFile(body []byte) (*fit.File, error) {
	fit, err := fit.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to decode object body: %v", err)
	}
	return fit, nil
}

func DecodeAltitude(storedValue uint32) float32 {
	return (float32(storedValue) / 5) - 500.0
}

func PrintFITFileDetails(fit *fit.File) {
	fmt.Println(fit.FileId.TimeCreated)
	fmt.Println(fit.FileId.GetProduct())
	fmt.Println(fit.Type())
}

// getFITFileActivity is a function that retrieves the activity from a FIT file.
// It takes a fit.File as input and returns the activity and an error.
// If the activity cannot be retrieved, it returns an error with a descriptive message.
func GetFITFileActivity(fit *fit.File) (*fit.ActivityFile, error) {
	activity, err := fit.Activity()
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %v", err)
	}
	return activity, nil
}
