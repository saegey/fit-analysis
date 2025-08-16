package myevent

import (
	"math"
	"time"

	"github.com/tormoder/fit"
)

// CalculateMaxAveragePowers calculates the best average power over specific intervals
func CalculateMaxAveragePowers(intervals []uint16, powers []uint16) map[uint16]int {
	bestPowers := make(map[uint16]int)

	for _, interval := range intervals {
		maxAvg := float64(0)
		windowSum := uint64(0)

		// Sliding window approach
		for i := uint16(0); i < uint16(len(powers)); i++ {
			if i < interval {
				windowSum += uint64(powers[i])
				if i == interval-1 {
					maxAvg = float64(windowSum) / float64(interval)
				}
			} else {
				windowSum += uint64(powers[i]) - uint64(powers[i-interval])
				avg := float64(windowSum) / float64(interval)
				if avg > maxAvg {
					maxAvg = avg
				}
			}
		}
		bestPowers[interval] = int(maxAvg)
	}

	return bestPowers
}

func CalcNormalizedPower(powers []uint16) float32 {
	segmentSums := make([]float64, 0)
	cleanPowers := make([]float64, len(powers))
	for i, p := range powers {
		if p == 0 {
			cleanPowers[i] = 0
		} else {
			cleanPowers[i] = float64(p)
		}
	}

	index := 0
	for index <= len(powers)-30 {
		segmentSum := 0.0
		for _, p := range cleanPowers[index : index+30] {
			segmentSum += p
		}
		segmentSums = append(segmentSums, segmentSum)
		index += 1
	}

	segmentTotal := 0.0
	for _, s := range segmentSums {
		segmentTotal += math.Pow(s/30, 4)
	}

	average := segmentTotal / float64(len(segmentSums))
	return float32(math.Pow(average, 1.0/4.0))
}

func isFinite(f float64) bool {
	return !math.IsInf(f, 0) && !math.IsNaN(f)
}

func CalcElevationGrades(coordinates [][]float64, distances []float32, elevations []float64) []float64 {
	grades := make([]float64, len(coordinates))
	for index := range coordinates {
		grade := 0.0
		if !math.IsNaN(grade) && isFinite(grade) {
			if index > 30 && index-30 < len(coordinates) {
				grade = float64(elevations[index]-elevations[index-30]) / float64(distances[index]-distances[index-30])
			}

			if !math.IsNaN(grade) && isFinite(grade) {
				grades[index] = math.Round(grade*1000) / 1000
			} else {
				grades[index] = 0
			}
		}
	}
	return grades
}

func GetMaxAveragePowerForInterval(records []*fit.RecordMsg, interval time.Duration) float64 {
	maxAverage := 0.0
	start := 0
	windowSum := uint64(0)

	for end := 0; end < len(records); end++ {
		// Move the start index up to maintain the interval window
		for records[end].Timestamp.Sub(records[start].Timestamp) > interval {
			windowSum -= uint64(records[start].Power)
			start++
		}
		// Add the current record's power to the window sum
		windowSum += uint64(records[end].Power)
		// Calculate the average for the current window
		currentWindowLength := end - start + 1
		average := float64(windowSum) / float64(currentWindowLength)
		if average > maxAverage {
			maxAverage = average
		}
	}

	return maxAverage
}
