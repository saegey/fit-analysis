package powercalc

func GenerateIntervals(maxTime int) []uint16 {
	baseIntervals := []uint16{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 20, 25, 30, 35, 40, 45, 50,
		55, 60, 70, 80, 90, 100, 110, 120, 180, 240, 300, 360, 420, 480, 540, 600,
		660, 720, 780, 840, 900, 960, 1020, 1080, 1140, 1200, 1500, 1800, 2100, 2400,
		2700, 3000, 3300, 3600, 5400,
	}

	for hour := 2; hour*3600 <= maxTime; hour++ {
		baseIntervals = append(baseIntervals, uint16(hour*3600))
	}
	baseIntervals = append(baseIntervals, uint16(maxTime-1))

	return baseIntervals
}
