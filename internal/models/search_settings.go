package models

type SearchSettings struct {
	Weights                  Weigts
	Genres                   []GenresMarked
	OldRating                []IsMarked
	Kind                     []IsMarked
	MinRating, MaxRating     float64
	MinDuration, MaxDuration int
	MinEpisodes, MaxEpisodes int
	MinYear, MaxYear         int
}

func (sets SearchSettings) GenresArr() GenresMarked {
	var arr = make(GenresMarked, 0)
	for _, genres := range sets.Genres {
		arr = append(arr, genres...)
	}
	return arr
}

func (sets SearchSettings) SwapGenre(name string) {
	for i, genres := range sets.Genres {
		for j := range genres {
			if sets.Genres[i][j].Name == name {
				sets.Genres[i][j].Marked = !sets.Genres[i][j].Marked
			}
		}
	}
}

func (sets SearchSettings) SwapKind(goal string) {
	for i, v := range sets.Kind {
		if v.Value == goal {
			sets.Kind[i].Marked = !v.Marked
		}
	}
}

func (sets SearchSettings) SwapOldRating(goal string) {
	for i, v := range sets.OldRating {
		if v.Value == goal {
			sets.OldRating[i].Marked = !v.Marked
		}
	}
}
