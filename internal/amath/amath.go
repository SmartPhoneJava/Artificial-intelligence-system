package amath

import "math"

type Pairs []float64

func (pairs *Pairs) Add(a, b, k float64) {
	if k < 0.01 {
		k = 1
	}
	*pairs = append(*pairs, k*a, k*b)
}

func (pairs *Pairs) AddInt(a, b int, k float64) {
	pairs.Add(float64(a), float64(b), k)
}

func (pairs *Pairs) AddString(a, b string, k float64) {
	if a != b {
		pairs.Add(1, -1, k)
	}
}

func (pairs *Pairs) AddBool(a, b bool, k float64) {
	if a != b {
		pairs.Add(1, -1, k)
	}
}

func (pairs *Pairs) AddSlice(a, b []string, k float64) int {
	var different = len(a) + len(b)
	for _, s1 := range a {
		for _, s2 := range b {
			if s1 == s2 {
				different -= 2
				break
			}
		}
	}
	if different != 0 {
		pairs.AddInt(different, 0, k)
	}
	return different
}

// Euclidean distance
// Евклидово расстояние (расстояние по прямой)
func (pairs Pairs) Euclidean() float64 {
	var summ, prev float64
	for i, p := range pairs {
		if i%2 == 0 {
			prev = p
		} else {
			summ += math.Pow(p-prev, 2)
		}
	}
	return math.Sqrt(summ)
}

// L1 distance
// Расстояние L1 (расстояние городских кварталов)
func (pairs Pairs) L1() float64 {
	var summ, prev float64
	for i, p := range pairs {
		if i%2 == 0 {
			prev = p
		} else {
			summ += math.Abs(p - prev)
		}
	}
	return summ
}

// Chess distance
// Расстояние Чебышёва (метрика шахматной доски)
func (pairs Pairs) Chess() float64 {
	var max, prev float64
	for i, p := range pairs {
		if i%2 == 0 {
			prev = p
		} else {
			summ := math.Abs(p - prev)
			if summ > max {
				max = summ
			}
		}
	}
	return max
}
