package simplify

type Stack []int

func (s *Stack) Push(v int) {
	*s = append(*s, v)
}

func (s *Stack) Pop() int {
	if len(*s) > 0 {
		ret := (*s)[len(*s)-1]
		*s = (*s)[0 : len(*s)-1]
		return ret
	} else {
		return 0
	}
}

//-------------------------------------------------------

func getSqDist(p1, p2 []float64) float64 {
	dx := p1[0] - p2[0]
	dy := p1[1] - p2[1]
	return dx*dx + dy*dy
}

func getSqSegDist(p, p1, p2 []float64) float64 {
	x := p1[0]
	y := p1[1]
	dx := p2[0] - x
	dy := p2[1] - y
	if dx != 0 || dy != 0 {
		t := ((p[0]-x)*dx + (p[1]-y)*dy) / (dx*dx + dy*dy)
		if t > 1 {
			x = p2[0]
			y = p2[1]
		} else if t > 0 {
			x += dx * t
			y += dy * t
		}
	}
	dx = p[0] - x
	dy = p[1] - y
	return dx*dx + dy*dy
}

// SimplifyRadialDistWithIndices simplifies points using radial distance and returns simplified points with indices.
func SimplifyRadialDistWithIndices(points [][]float64, sqTolerance float64) ([][]float64, []int) {
	prevPoint := points[0]
	newPoints := [][]float64{prevPoint}
	indices := []int{0}
	var point []float64
	for i := 1; i < len(points); i++ {
		point = points[i]
		if getSqDist(point, prevPoint) > sqTolerance {
			newPoints = append(newPoints, point)
			indices = append(indices, i)
			prevPoint = point
		}
	}
	if !ComparePoints(prevPoint, point) {
		newPoints = append(newPoints, point)
		indices = append(indices, len(points)-1)
	}
	return newPoints, indices
}

// SimplifyDouglasPeuckerWithIndices simplifies points using the Douglas-Peucker algorithm and returns simplified points with indices.
func SimplifyDouglasPeuckerWithIndices(points [][]float64, sqTolerance float64) ([][]float64, []int) {
	var l = len(points)
	if l == 0 {
		return nil, nil
	}
	markers := make([]int, l)
	first := 0
	last := l - 1
	var stack Stack
	var newPoints [][]float64
	var indices []int
	i, index := 0, 0
	maxSqDist, sqDist := float64(0), float64(0)
	markers[first], markers[last] = 1, 1

	stack.Push(first)
	stack.Push(last)

	for len(stack) > 0 {
		last = stack.Pop()
		first = stack.Pop()

		maxSqDist = 0
		index = first + 1

		for i = first + 1; i < last; i++ {
			sqDist = getSqSegDist(points[i], points[first], points[last])
			if sqDist > maxSqDist {
				index = i
				maxSqDist = sqDist
			}
		}
		if maxSqDist > sqTolerance {
			markers[index] = 1
			stack.Push(first)
			stack.Push(index)
			stack.Push(index)
			stack.Push(last)
		}
	}

	for i = 0; i < l; i++ {
		if markers[i] != 0 {
			newPoints = append(newPoints, points[i])
			indices = append(indices, i)
		}
	}
	return newPoints, indices
}

// SimplifyWithIndices simplifies points and returns the simplified points along with their original indices.
// If highestQuality is true, radial distance simplification is skipped.
func SimplifyWithIndices(points [][]float64, tolerance float64, highestQuality bool) ([][]float64, []int) {
	if len(points) <= 2 {
		indices := make([]int, len(points))
		for i := range points {
			indices[i] = i
		}
		return points, indices
	}
	sqTolerance := tolerance * tolerance
	var _points [][]float64
	var preIndices []int
	if highestQuality {
		_points = points
		preIndices = make([]int, len(points))
		for i := range points {
			preIndices[i] = i
		}
	} else {
		_points, preIndices = SimplifyRadialDistWithIndices(points, sqTolerance)
	}
	simplifiedPoints, indices := SimplifyDouglasPeuckerWithIndices(_points, sqTolerance)

	// Map indices back to original points
	finalIndices := make([]int, len(indices))
	for i, idx := range indices {
		finalIndices[i] = preIndices[idx]
	}

	return simplifiedPoints, finalIndices
}

// -------------------------------
func CompareSlices(p1, p2 [][]float64) bool {
	if len(p1) == len(p2) {
		for i := range p1 {
			if !ComparePoints(p1[i], p2[i]) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func ComparePoints(p1, p2 []float64) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}
