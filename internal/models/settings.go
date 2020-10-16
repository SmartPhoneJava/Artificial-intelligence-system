package models

type Weigts struct {
	Kind     float64 `json:"kind"`
	Score    float64 `json:"score"`
	Episodes float64 `json:"episodes"`
	Duration float64 `json:"duration"`
	Rating   float64 `json:"rating"`
	Year     float64 `json:"year"`
	Ongoing  float64 `json:"ongoing"`
	Studio   float64 `json:"studio"`
	Genre    float64 `json:"genre"`
}

func DefaultWeigts() Weigts {
	return Weigts{
		Kind:     1,
		Score:    1,
		Episodes: 3,
		Duration: 2,
		Rating:   5,
		Year:     0.05,
		Ongoing:  5,
		Studio:   3,
		Genre:    10,
	}
}
