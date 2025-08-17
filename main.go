package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"lambda/dynamo"
	"lambda/fitHelper"
	"math"
	"os"
)

// Output is the JSON struct we print to stdout containing only the requested fields.
type Output struct {
	HeartAnalysis         dynamo.HeartAnalysis `json:"HeartAnalysis"`
	ElevationGain         float32              `json:"ElevationGain"`
	StoppedTime           int                  `json:"StoppedTime"`
	ElapsedTime           int                  `json:"ElapsedTime"`
	NormalizedPower       float32              `json:"NormalizedPower"`
	PowerAnalysis         dynamo.PowerAnalysis `json:"PowerAnalysis"`
	SimplifiedCoordinates [][]float64          `json:"SimplifiedCoordinates"`
	SimplifiedDistances   []float32            `json:"SimplifiedDistances"`
	SimplifiedElevations  []float64            `json:"SimplifiedElevations"`
	PowerZoneBuckets      []int                `json:"PowerZoneBuckets,omitempty"`
	PowerZones            []PowerZone          `json:"PowerZones,omitempty"`
}

// PowerZone is the computed zone boundaries based on FTP
type PowerZone struct {
	Zone      int    `json:"zone"`
	Title     string `json:"title"`
	PowerLow  *int   `json:"powerLow"`
	PowerHigh *int   `json:"powerHigh"`
}

func main() {
	fitPath := flag.String("fit", "", "Path to the .fit file to process")
	ftp := flag.Int("ftp", 0, "Optional FTP value to compute power zones/buckets")
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
		Activity: activity,
	}

	processedData, err := ProcessActivityRecords(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process activity: %v\n", err)
		os.Exit(1)
	}

	out := Output{
		HeartAnalysis:         processedData.HeartResults,
		ElevationGain:         processedData.TotalElevationGain,
		StoppedTime:           processedData.StoppedTime,
		ElapsedTime:           processedData.ElapsedTime,
		NormalizedPower:       processedData.NormalizedPower,
		PowerAnalysis:         processedData.PowerResults,
		SimplifiedCoordinates: processedData.SimplifiedCoordinates,
		SimplifiedDistances:   processedData.SimplifiedDistances,
		SimplifiedElevations:  processedData.SimplifiedElevations,
	}

	// If FTP provided, calculate power zones and buckets
	if *ftp > 0 {
		zones := calcPowerZones(*ftp)

		// Compute time-based buckets (seconds spent in each zone) using merged simplified timeseries
		buckets := make([]int, len(zones))
		merged := processedData.MergedData
		if len(merged) > 0 {
			for i := 0; i < len(merged); i++ {
				p := merged[i].Power

				// find zone for this power
				zoneIdx := -1
				for j := len(zones) - 1; j >= 0; j-- {
					if zones[j].PowerLow != nil && int(p) >= *zones[j].PowerLow {
						zoneIdx = j
						break
					}
				}
				if zoneIdx == -1 {
					continue
				}

				// compute duration until next simplified point (or until elapsed time for last point)
				var dt float64
				if i < len(merged)-1 {
					dt = merged[i+1].Time - merged[i].Time
				} else {
					dt = float64(processedData.ElapsedTime) - merged[i].Time
				}
				if dt < 0 {
					dt = 0
				}
				buckets[zoneIdx] += int(math.Round(dt))
			}
		}

		out.PowerZones = zones
		out.PowerZoneBuckets = buckets
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}

// calcPowerZones returns the power zones based on FTP, mirroring the JS function.
func calcPowerZones(ftp int) []PowerZone {
	zonesPercent := []PowerZone{
		{Zone: 0, Title: "Not pedaling", PowerLow: intPtr(0), PowerHigh: intPtr(0)},
		{Zone: 1, Title: "Active Recovery", PowerLow: intPtr(1), PowerHigh: intPtr(56)},
		{Zone: 2, Title: "Endurance", PowerLow: intPtr(56), PowerHigh: intPtr(76)},
		{Zone: 3, Title: "Tempo", PowerLow: intPtr(76), PowerHigh: intPtr(91)},
		{Zone: 4, Title: "Threshold", PowerLow: intPtr(91), PowerHigh: intPtr(106)},
		{Zone: 5, Title: "VO2max", PowerLow: intPtr(106), PowerHigh: intPtr(121)},
		{Zone: 6, Title: "Anaerobic Capacity", PowerLow: intPtr(121), PowerHigh: nil},
	}

	res := make([]PowerZone, len(zonesPercent))
	for i, z := range zonesPercent {
		var low *int
		var high *int
		if z.PowerLow != nil {
			v := int((float64(*z.PowerLow) / 100.0) * float64(ftp))
			low = &v
		}
		if z.PowerHigh != nil {
			v := int((float64(*z.PowerHigh) / 100.0) * float64(ftp))
			high = &v
		}
		res[i] = PowerZone{
			Zone:      z.Zone,
			Title:     z.Title,
			PowerLow:  low,
			PowerHigh: high,
		}
	}
	return res
}

func intPtr(i int) *int { return &i }

// calcPowerZoneBuckets returns counts of powers falling into each zone (same logic as JS)
func calcPowerZoneBuckets(zones []PowerZone, powers []uint16) []int {
	buckets := make([]int, len(zones))
	for _, p := range powers {
		// ignore zero-power samples (likely missing or coasting)
		// if p == 0 {
		// 	continue
		// }
		for i := len(zones) - 1; i >= 0; i-- {
			z := zones[i]
			if z.PowerLow != nil && int(p) >= *z.PowerLow {
				buckets[i] += 1
				break
			}
		}
	}
	return buckets
}
