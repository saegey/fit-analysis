package main

import (
	"math"
	"os"
	"testing"

	"github.com/tormoder/fit"
)

func TestWahooProcessActivityRecords(t *testing.T) {
	// Load the FIT file from the testdata directory
	file, err := os.Open("testdata/Rapha_WTFNB_100_.fit")
	if err != nil {
		t.Fatalf("Failed to open FIT file: %v", err)
	}
	defer file.Close()

	// Decode the FIT file
	fitFile, err := fit.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode FIT file: %v", err)
	}

	// Ensure the file is an activity file
	activityFile, err := fitFile.Activity()
	if err != nil {
		t.Fatalf("Failed to get activity data: %v", err)
	}

	// Create mock options for testing
	opts := ProcessActivityOptions{
		Svc:         nil, // Provide mock DynamoDB API implementation or nil for this test
		Activity:    activityFile,
		PostId:      nil,
		Bucket:      "test-bucket",
		IdentityId:  "test-identity",
		GpxFileName: "test.gpx",
	}

	// Call the function and check the result
	result, err := ProcessActivityRecords(opts)
	if err != nil {
		t.Fatalf("Processing activity failed: %v", err)
	}

	// Perform assertions on the returned result
	if result.TotalDistance != (110.087810) {
		t.Errorf("Expected total distance to be greater than 0, got %f", result.TotalDistance)
	}

	// Print the result to verify the output manually for now
	// jsonResult, err := json.MarshalIndent(result, "", "  ")
	// if err != nil {
	// 	t.Fatalf("Failed to marshal result to JSON: %v", err)
	// }
	// fmt.Printf("Processed Activity Data: %s\n", jsonResult)

	if result.NormalizedPower != 143.667770 {
		t.Errorf("Expected normalized power to be greater than 0, got %f", result.NormalizedPower)
	}

	if result.ElapsedTime != 24133 {
		t.Errorf("Expected elapsed time to be 10800, got %d", result.ElapsedTime)
	}

	if result.StoppedTime != 7167 {
		t.Errorf("Expected stopped time to be 20772, got %d", result.StoppedTime)
	}

	if result.TotalElevationGain != 1247.762207 {
		t.Errorf("Expected TotalElevationGain to be 1247.762207, got %f", result.TotalElevationGain)
	}

	if len(result.PowerResults) != 61 {
		t.Errorf("Expected power results length to be %d, got %d", len(result.PowerResults), len(result.PowerResults))
	}

	// Compare each key-value pair
	for key, expectedValue := range result.PowerResults {
		// fmt.Println(key, expectedValue)
		if resultValue, ok := result.PowerResults[key]; !ok {
			t.Errorf("Expected key %d in power results, but it was missing", key)
		} else if resultValue != expectedValue {
			t.Errorf("Expected power result for key %d to be %d, got %d", key, expectedValue, resultValue)
		}
	}

	if len(result.CadenceResults) != 61 {
		t.Errorf("Expected CadenceResults length to be %d, got %d", len(result.CadenceResults), len(result.CadenceResults))
	}

	// Compare each key-value pair
	for key, expectedValue := range result.CadenceResults {
		// fmt.Println(key, expectedValue)
		if resultValue, ok := result.CadenceResults[key]; !ok {
			t.Errorf("Expected key %d in power results, but it was missing", key)
		} else if resultValue != expectedValue {
			t.Errorf("Expected power result for key %d to be %d, got %d", key, expectedValue, resultValue)
		}
	}

	if len(result.HeartResults) != 61 {
		t.Errorf("Expected HeartResults length to be %d, got %d", len(result.HeartResults), len(result.HeartResults))
	}

	// Compare each key-value pair
	for key, expectedValue := range result.HeartResults {
		// fmt.Println(key, expectedValue)
		if resultValue, ok := result.HeartResults[key]; !ok {
			t.Errorf("Expected key %d in power results, but it was missing", key)
		} else if resultValue != expectedValue {
			t.Errorf("Expected power result for key %d to be %d, got %d", key, expectedValue, resultValue)
		}
	}

	if len(result.TempResults) != 61 {
		t.Errorf("Expected temperature results length to be %d, got %d", len(result.TempResults), len(result.TempResults))
	}

	// Compare each key-value pair
	for key, expectedValue := range result.TempResults {
		// fmt.Println(key, expectedValue)
		if resultValue, ok := result.TempResults[key]; !ok {
			t.Errorf("Expected key %d in power results, but it was missing", key)
		} else if resultValue != expectedValue {
			t.Errorf("Expected power result for key %d to be %d, got %d", key, expectedValue, resultValue)
		}
	}

	if len(result.SimplifiedCoordinates) != 2300 {
		t.Fatalf("Expected simplified coordinates to be non-nil, got %v", len(result.SimplifiedCoordinates))
	}

	// Loop through the simplified coordinates to check for NaN values
	for i, coord := range result.SimplifiedCoordinates {
		if len(coord) != 3 {
			t.Fatalf("Expected each coordinate to have 3 values (lat, long, altitude), but got %d at index %d", len(coord), i)
		}

		lat, long, altitude := coord[0], coord[1], coord[2]

		if math.IsNaN(lat) || math.IsNaN(long) || math.IsNaN(altitude) {
			t.Errorf("Found NaN value at index %d: [%f, %f, %f]", i, lat, long, altitude)
		}
	}

	if result.MergedData == nil {
		t.Fatalf("Expected MergedData to be non-nil, got nil")
	}

	// Check if MergedData contains any entries
	if len(result.MergedData) == 0 {
		t.Fatalf("Expected MergedData to contain data, but it is empty")
	}

	if len(result.MergedData) != 2300 {
		t.Errorf("Merged data should have length of %d", len(result.MergedData))
	}

	// Loop through each item in the MergedData slice
	for i, data := range result.MergedData {
		// Check for non-NaN values
		if math.IsNaN(data.Distance) {
			t.Errorf("Found NaN for Distance at index %d", i)
		}

		if math.IsNaN(float64(data.Time)) {
			t.Errorf("Found NaN for Time at index %d", i)
		}

		if math.IsNaN(float64(data.Elevation)) {
			t.Errorf("Found NaN for Elevation at index %d", i)
		}

		if math.IsNaN(data.Grade) {
			t.Errorf("Found NaN for Grade at index %d", i)
		}

		// Add other checks depending on what you expect for each field
		// e.g., checking for valid ranges for distance, time, etc.
		if data.Distance < 0 {
			t.Errorf("Expected non-negative Distance, got %f at index %d", data.Distance, i)
		}

		if data.Distance >= 42949672.000000 {
			t.Errorf("Expected distance to be less than 42949672, got %f at index %d", data.Distance, i)
		}
	}

}
