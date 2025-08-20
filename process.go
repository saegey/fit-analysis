package main

import (
	"fmt"
	"lambda/S3helper"
	"lambda/dynamo"
	"lambda/fitHelper"
	"lambda/myevent"
	powercalc "lambda/powerCalc"
	"lambda/simplify"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tormoder/fit"
)

func convertToUint16Slice(slice []uint8) []uint16 {
	result := make([]uint16, len(slice))
	for i, v := range slice {
		result[i] = uint16(v)
	}
	return result
}

func convertToInt8Slice(slice []int8) []uint8 {
	result := make([]uint8, len(slice))
	for i, v := range slice {
		result[i] = uint8(v)
	}
	return result
}

type ProcessActivityOptions struct {
	Activity    *fit.ActivityFile
	PostId      *string
	Bucket      string
	IdentityId  string
	GpxFileName string
}

type ProcessedActivityData struct {
	SimplifiedCoordinates [][]float64
	SimplifiedDistances   []float32
	SimplifiedElevations  []float64
	MergedData            []S3helper.MergedDataItem
	S3Key                 string
	TempResults           dynamo.TempAnalysis
	CadenceResults        dynamo.CadenceAnalysis
	TotalDistance         float64
	HeartResults          dynamo.HeartAnalysis
	TotalElevationGain    float32
	StoppedTime           int
	ElapsedTime           int
	NormalizedPower       float32
	PowerResults          dynamo.PowerAnalysis
	RawPowers             []uint16
}

func ProcessActivityRecords(opts ProcessActivityOptions) (*ProcessedActivityData, error) {
	// Detect manufacturer (Wahoo = 89)
	// Variables for calculations and data storage
	var totalPower int
	var count int
	var powers []uint16
	var cads []uint8
	var temps []int8
	var coordinates [][]float64
	var distances []float32
	var elevations []float64
	var hearts []uint8

	// Coordinate-aligned slices to keep data in sync with `coordinates` entries
	var coordElevations []float64
	var coordPowers []uint16
	var coordHearts []uint8
	var coordDistances []float32

	var totalElevationGain float32
	var previousAltitude float32
	first := true
	var stoppedTime time.Duration
	var activity = opts.Activity

	// Determine device manufacturer from DeviceInfos (Wahoo Fitness = 89)
	isWahoo := false
	if activity != nil {
		for _, di := range activity.DeviceInfos {
			if di != nil {
				if uint16(di.Manufacturer) == 89 { // Wahoo Fitness per FIT profile
					isWahoo = true
					break
				}
			}
		}
	}

	// Loop through each record in the activity file
	for i, record := range opts.Activity.Records {
		power := record.Power
		if i > 0 {
			prevRecord := activity.Records[i-1]
			currRecord := activity.Records[i]

			// If the distance hasn't changed, assume the time is "stopped"
			if currRecord.Distance == prevRecord.Distance {
				stoppedTime += currRecord.Timestamp.Sub(prevRecord.Timestamp)
			}

			// Store power data
			if power != 65535 {
				totalPower += int(power)
				powers = append(powers, power)
			} else {
				powers = append(powers, 0)
			}

			// Store cadence data
			if record.Cadence != 255 {
				cads = append(cads, uint8(record.Cadence))
			} else {
				cads = append(cads, 0)
			}

			// Store temperature data
			if record.Temperature != 0 {
				temps = append(temps, record.Temperature)
			}

			// Store heart rate data
			hearts = append(hearts, record.HeartRate)

			// Store coordinate data
			if record.PositionLat.Degrees() != 0 && record.PositionLong.Degrees() != 0 {
				lat := float64(record.PositionLat.Degrees())
				long := float64(record.PositionLong.Degrees())
				altitude := float64(record.EnhancedAltitude)

				// Ensure no NaN values are added
				if !math.IsNaN(lat) && !math.IsNaN(long) && !math.IsNaN(altitude) && altitude != 4294967295 {
					// compute decoded elevation in feet (matching previous behavior)
					decodedAlt := fitHelper.DecodeAltitude(record.EnhancedAltitude)
					elevFt, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(decodedAlt)*3.28084), 64)
					coordinates = append(coordinates, []float64{long, lat, altitude})
					coordElevations = append(coordElevations, elevFt)
					if power != 65535 {
						coordPowers = append(coordPowers, power)
					} else {
						coordPowers = append(coordPowers, 0)
					}
					coordHearts = append(coordHearts, record.HeartRate)
					if record.Distance != 4294967295 {
						coordDistances = append(coordDistances, (float32(record.Distance) / 100))
					} else {
						coordDistances = append(coordDistances, 0)
					}
				}
			}

			// Store distance data
			// 4,294,967,295
			if record.Distance != 4294967295 {
				distances = append(distances, (float32(record.Distance) / 100)) // convert to meters
			}
			count++

			// Calculate elevation gain: Wahoo -> EnhancedAltitude, Others -> Altitude
			var rawAltitude32 uint32
			var decodedAltitude float32
			if isWahoo {
				// Prefer Altitude for Wahoo; fallback to EnhancedAltitude
				if record.Altitude != 65535 && record.Altitude != 0 {
					rawAltitude32 = uint32(record.Altitude)
					decodedAltitude = fitHelper.DecodeAltitude(rawAltitude32)
				} else if record.EnhancedAltitude != 4294967295 && record.EnhancedAltitude != 0 { // fallback
					rawAltitude32 = record.EnhancedAltitude
					decodedAltitude = fitHelper.DecodeAltitude(rawAltitude32)
				} else {
					continue
				}
			} else {
				// Prefer EnhancedAltitude for non-Wahoo; fallback to Altitude
				if record.EnhancedAltitude != 4294967295 && record.EnhancedAltitude != 0 {
					rawAltitude32 = record.EnhancedAltitude
					decodedAltitude = fitHelper.DecodeAltitude(rawAltitude32)
				} else if record.Altitude != 65535 && record.Altitude != 0 { // fallback
					rawAltitude32 = uint32(record.Altitude)
					decodedAltitude = fitHelper.DecodeAltitude(rawAltitude32)
				} else {
					continue
				}
			}

			elevation, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(decodedAltitude)*3.28084), 32)
			elevations = append(elevations, elevation)

			if !first {
				if decodedAltitude > previousAltitude {
					gain := decodedAltitude - previousAltitude
					totalElevationGain += gain
				}
			} else {
				first = false
			}

			previousAltitude = decodedAltitude
		}
	}

	// Check if any records were found
	// If there are no records at all, bail out early to avoid panics when indexing
	if len(activity.Records) == 0 {
		fmt.Println("Activity contains no records")
		return nil, fmt.Errorf("activity contains no records")
	}

	// Check if any processed records were found
	if count == 0 {
		fmt.Println("No valid records found after filtering")
		return nil, fmt.Errorf("no records found")
	}

	// Calculate total distance
	totalDistance := (float64(activity.Records[len(activity.Records)-1].Distance) / 100) / 1000
	normalizedPower := myevent.CalcNormalizedPower(powers)

	// Calculate best powers for different time intervals
	var timeIntervals = powercalc.GenerateIntervals(len(activity.Records))

	var powerResults = myevent.CalculateMaxAveragePowers(timeIntervals, powers)
	var cadenceResults = myevent.CalculateMaxAveragePowers(timeIntervals, convertToUint16Slice(cads))
	var tempResults = myevent.CalculateMaxAveragePowers(timeIntervals, convertToUint16Slice([]uint8(convertToInt8Slice(temps))))
	var heartResults = myevent.CalculateMaxAveragePowers(timeIntervals, convertToUint16Slice(hearts))
	var elapsedTime = activity.Records[len(activity.Records)-1].Timestamp.Sub(activity.Records[0].Timestamp)

	// Simplify coordinates and get indices
	simplifiedCoordinates, indices := simplify.SimplifyWithIndices(coordinates, 0.00001, true)

	// Use indices to extract corresponding data
	simplifiedElevations := make([]float64, len(indices))
	simplifiedPowers := make([]uint16, len(indices))
	simplifiedHearts := make([]uint8, len(indices))
	simplifiedDistances := make([]float32, len(indices))

	// Use coordinate-aligned slices to avoid mismatches
	for i, idx := range indices {
		if idx < 0 || idx >= len(coordElevations) {
			// safety: skip out-of-range indices
			continue
		}
		simplifiedElevations[i] = coordElevations[idx]
		simplifiedPowers[i] = coordPowers[idx]
		simplifiedHearts[i] = coordHearts[idx]
		simplifiedDistances[i] = coordDistances[idx]
	}

	// Now you can create MergedData using the simplified data
	mergedData := make([]S3helper.MergedDataItem, len(indices))
	for i := range indices {
		var grade float64
		if i == 0 {
			grade = 0
		} else {
			elevationChange := simplifiedElevations[i] - simplifiedElevations[i-1]
			distanceChange := float64(simplifiedDistances[i] - simplifiedDistances[i-1])
			if distanceChange != 0 {
				grade = elevationChange / distanceChange
			}
		}

		var t float64
		idx := indices[i]
		if idx >= 0 && idx < len(activity.Records) {
			t = activity.Records[idx].Timestamp.Sub(activity.Records[0].Timestamp).Seconds()
		}

		mergedData[i] = S3helper.MergedDataItem{
			Power:     simplifiedPowers[i],
			Distance:  float64(simplifiedDistances[i]),
			Time:      t,
			Elevation: float32(simplifiedElevations[i]),
			HeartRate: simplifiedHearts[i],
			Grade:     grade,
		}
	}

	// Generate S3 key
	s3key := fmt.Sprintf("timeseries/%s.json", uuid.New().String())

	// Return the processed data
	return &ProcessedActivityData{
		SimplifiedCoordinates: simplifiedCoordinates,
		SimplifiedDistances:   simplifiedDistances,
		SimplifiedElevations:  simplifiedElevations,
		MergedData:            mergedData,
		S3Key:                 s3key,
		TempResults:           tempResults,
		CadenceResults:        cadenceResults,
		TotalDistance:         totalDistance,
		HeartResults:          heartResults,
		TotalElevationGain:    totalElevationGain,
		StoppedTime:           int(stoppedTime.Seconds()),
		ElapsedTime:           int(elapsedTime.Seconds()),
		NormalizedPower:       normalizedPower,
		PowerResults:          powerResults,
	}, nil
}
