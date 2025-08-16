package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"lambda/dynamo"
	"lambda/fitHelper"
	"os"
)

// Output is the JSON struct we print to stdout containing only the requested fields.
type Output struct {
	HeartAnalysis   dynamo.HeartAnalysis `json:"HeartAnalysis"`
	ElevationGain   float32              `json:"ElevationGain"`
	StoppedTime     int                  `json:"StoppedTime"`
	ElapsedTime     int                  `json:"ElapsedTime"`
	NormalizedPower float32              `json:"NormalizedPower"`
	PowerAnalysis   dynamo.PowerAnalysis `json:"PowerAnalysis"`
}

func main() {
	fitPath := flag.String("fit", "", "Path to the .fit file to process")
	flag.Parse()

	if *fitPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: ./processFitFile --fit <file.fit>")
		os.Exit(1)
	}

	body, err := ioutil.ReadFile(*fitPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}

	fitFile, err := fitHelper.DecodeFITFile(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode FIT file: %v\n", err)
		os.Exit(1)
	}

	activity, err := fitHelper.GetFITFileActivity(fitFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get activity: %v\n", err)
		os.Exit(1)
	}

	opts := ProcessActivityOptions{
		Svc:      nil,
		Activity: activity,
	}

	processedData, err := ProcessActivityRecords(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process activity: %v\n", err)
		os.Exit(1)
	}

	out := Output{
		HeartAnalysis:   processedData.HeartResults,
		ElevationGain:   processedData.TotalElevationGain,
		StoppedTime:     processedData.StoppedTime,
		ElapsedTime:     processedData.ElapsedTime,
		NormalizedPower: processedData.NormalizedPower,
		PowerAnalysis:   processedData.PowerResults,
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}
