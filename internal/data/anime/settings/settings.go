package settings

type AnimeSettings struct {
	Kind               float64
	Score              float64
	Episodes, Duration float64
	Rating             float64
	Year               float64
	Ongoing            float64
	Studio, Genre      float64
}

func NewAnimeSettings() AnimeSettings {
	return AnimeSettings{
		Kind:     1,
		Score:    1,
		Episodes: 3,
		Duration: 2,
		Rating:   5,
		Year:     6,
		Ongoing:  5,
		Studio:   3,
		Genre:    10,
	}
}
