package amath

import (
	"errors"
	"math"

	"gonum.org/v1/gonum/stat"
)

// type Matrix [][]float64

// func NewMatrix(pairs []Pairs) {
// 	matrix := make([][]float64, len(pairs))
// 	for i, p := range pairs {
// 		matrix[i] = make([]float64, len(p)/2)
// 		for j := 0; j <= len(p)/2; j += 2 {
// 			matrix[i][j/2] = p[j]
// 		}
// 	}
// }

type Pairs []float64

func NewPairs(a, b []float64) (Pairs, error) {
	if len(a) != len(b) {
		return Pairs{}, errors.New("len(a) not equal len(b)")
	}
	pairs := make([]float64, len(a)*2)
	for i := 0; i < len(a); i += 1 {
		pairs[i*2] = a[i]
		pairs[i*2+1] = b[i]
	}

	return pairs, nil
}

func (pairs Pairs) TwoVectors() ([]float64, []float64) {
	var (
		arr1 = make([]float64, len(pairs)/2)
		arr2 = make([]float64, len(pairs)/2)
	)
	for i := 1; i < len(pairs); i += 2 {
		arr1[(i-1)/2] = pairs[i-1]
		arr2[(i-1)/2] = pairs[i]
	}

	return arr1, arr2
}

func (pairs *Pairs) Add(a, b, k float64) {
	if k < 0.01 {
		k = 1
	}
	//fmt.Println("!!!", k*a, k*b)
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

// Diff difference
// Количество отличий
func (pairs Pairs) Diff() float64 {
	var summ, prev float64
	for i, p := range pairs {
		if i%2 == 0 {
			prev = p
		} else {
			if prev != p {
				summ++
			}
		}
	}
	return summ
}

// Correlation
func (pairs Pairs) Correlation() float64 {
	a, b := pairs.TwoVectors()
	return stat.Correlation(a, b, nil)
}
