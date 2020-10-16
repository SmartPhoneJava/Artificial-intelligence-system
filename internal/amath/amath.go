package amath

import (
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

func NewPairs(a, b []float64) Pairs {
	if len(a) != len(b) {
		panic("amath: slice length mismatch")
	}
	pairs := make([]float64, len(a)*2)
	for i := 0; i < len(a); i++ {
		pairs[i*2] = a[i]
		pairs[i*2+1] = b[i]
	}

	return pairs
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
		return
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

// Получает на вход число не совпавших элементов в слайсах
// вовзращает значение различия двух слайсов
type CompareSlices func(int) int

// Например:
// линейная: получив 5 вернём 5
func Linear(n int) int {
	return n
}

// квадратная: получив 5 вернём 55(1+4+9+16+25) - чем больше различий, тем сильнее различаются массивы
func Square(n int) int {
	var square = 0
	for i := 1; i <= n; i++ {
		square += i * i
	}
	return square
}

func LinearF(s, n, d float64) float64 {
	if n == 0 {
		return 0.01
	}
	var square = s
	for i := d; i <= n; i += d {
		square += i
	}
	return square
}

type SliceToString interface {
	Names() []string
}

func (pairs *Pairs) AddSlice(
	arr1, arr2 SliceToString,
	compareSlices CompareSlices,
	k float64,
) {
	var (
		a         = arr1.Names()
		b         = arr2.Names()
		different = len(a) + len(b)
	)

	for _, s1 := range a {
		for _, s2 := range b {
			if s1 == s2 {
				different -= 2
				break
			}
		}
	}
	if different != 0 {
		pairs.AddInt(compareSlices(different/2), 0, k)
	}
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
	return math.Abs(stat.Correlation(a, b, nil))
}

// Pearson correlation
func (pairs Pairs) PCorrelation() float64 {
	x, y := pairs.TwoVectors()

	if len(x) != len(y) {
		panic("amath: slice length mismatch")
	}
	xu := stat.Mean(x, nil)
	yu := stat.Mean(y, nil)
	var (
		sxx           float64
		syy           float64
		xcompensation float64
		ycompensation float64
	)

	for i, xv := range x {
		yv := y[i]
		xd := xv - xu
		yd := yv - yu
		sxx += xd * xd
		syy += yd * yd
		xcompensation += xd
		ycompensation += yd
	}

	return xcompensation * ycompensation / math.Sqrt(sxx*syy)
}
