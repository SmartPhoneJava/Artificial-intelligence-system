package amath

type Distances []float64

func NewDistances(n int) Distances {
	return make([]float64, n)
}

func (d Distances) Add(f float64, i int) { d[i] = f }
func (d Distances) Len() int             { return len(d) }
func (d Distances) Less(i, j int) bool   { return d[i] < d[j] }
func (d Distances) Swap(i, j int)        { d[i], d[j] = d[j], d[i] }
