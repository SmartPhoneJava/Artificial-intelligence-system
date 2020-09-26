package models

import "gonum.org/v1/gonum/floats"

type Anime struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	Russian       string `json:"russian"`
	URL           string `json:"url"`
	Status        string `json:"status"`
	EpisodesAired int    `json:"episodes_aired"`
	AiredOn       string `json:"aired_on"`
	ReleasedOn    string `json:"released_on"`
	Score         string `json:"score"`
	Rating        string `json:"rating"` //!++

	Kind     string  `json:"kind"` //!++
	Scoref   float64 //!+
	Episodes int     `json:"episodes"` //!+
	Duration int     `json:"duration"` //!++
	RatingI  int     //!++
	Year     int     //!++
	Ongoing  bool    `json:"ongoing"`

	Studios Studios  `json:"studios"` //!+++
	Genres  Genres   `json:"genres"`  //!+++
	Branch  []string `json:"branch"`  //!+++

	E, M, K, C, T, D float64
}

func (anime *Anime) WithСoefficients(
	E, M, K, C, T, D float64,
) {
	anime.E = floats.Round(E, 3)
	anime.M = floats.Round(M, 3)
	anime.K = floats.Round(K, 3)
	anime.C = floats.Round(C, 4)
	anime.T = floats.Round(T, 3)
	anime.D = floats.Round(D, 3)
}
