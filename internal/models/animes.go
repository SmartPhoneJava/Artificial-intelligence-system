package models

import (
	"fmt"
	"shiki/internal/amath"
)

type Animes []Anime

type IsMarked struct {
	Value  string
	Marked bool
}

func (a Animes) Pointers() []*Anime {
	var animes = make([]*Anime, len(a))
	for i := range a {
		animes[i] = &a[i]
	}
	return animes
}

func (a Animes) Top(n int) Animes {
	if len(a) < n {
		return a
	}
	return a[:n]
}

func (a Animes) Copy() Animes {
	var newArr = make([]Anime, len(a))
	copy(newArr, a)
	return newArr
}

func countDiff(startValue, weight, denominator float64) float64 {
	return amath.LinearF(startValue, (10-weight)/denominator, 1./denominator-0.001)
}

func printCommentDiff(startValue, denominator float64) {
	for b := 1; b <= 10; b++ {
		c := float64(10-b) / denominator
		count := startValue
		d := 1./denominator - 0.001

		for a := d; a <= c; a += d {
			count += a

		}
		fmt.Printf("//вес - %d, шаг - %.2f\n", b, count)
	}
}

func (a Animes) Filter(settings *SearchSettings) (Animes, string) {
	var (
		minD = settings.MinDuration
		maxD = settings.MaxDuration
		minE = settings.MinEpisodes
		maxE = settings.MaxEpisodes
		minY = settings.MinYear
		maxY = settings.MaxYear
		minR = settings.MinRating
		maxR = settings.MaxRating

		//вес - 1, шаг - 4.55
		//вес - 2, шаг - 3.64
		//вес - 3, шаг - 2.83
		//вес - 4, шаг - 2.12
		//вес - 5, шаг - 1.52
		//вес - 6, шаг - 1.01
		//вес - 7, шаг - 0.61
		//вес - 8, шаг - 0.30
		//вес - 9, шаг - 0.10
		//вес - 10, шаг - 0.00
		diffD = countDiff(0, settings.Weights.Duration, 9)
		plusD = 0.

		//вес - 1, шаг - 1.92
		//вес - 2, шаг - 1.64
		//вес - 3, шаг - 1.15
		//вес - 4, шаг - 0.94
		//вес - 5, шаг - 0.75
		//вес - 6, шаг - 0.45
		//вес - 7, шаг - 0.33
		//вес - 8, шаг - 0.17
		//вес - 9, шаг - 0.12
		//вес - 10, шаг - 0.00
		diffE = countDiff(0.1, settings.Weights.Episodes, 30)
		plusE = 0.

		//вес - 1, шаг - 7.05
		//вес - 2, шаг - 5.64
		//вес - 3, шаг - 4.39
		//вес - 4, шаг - 3.29
		//вес - 5, шаг - 2.35
		//вес - 6, шаг - 1.57
		//вес - 7, шаг - 0.94
		//вес - 8, шаг - 0.47
		//вес - 9, шаг - 0.16
		//вес - 10, шаг - 0.00
		diffY = countDiff(0, settings.Weights.Year, 6)
		plusY = 0.

		//вес - 1, шаг - 0.31
		//вес - 2, шаг - 0.26
		//вес - 3, шаг - 0.20
		//вес - 4, шаг - 0.16
		//вес - 5, шаг - 0.09
		//вес - 6, шаг - 0.06
		//вес - 7, шаг - 0.03
		//вес - 8, шаг - 0.02
		//вес - 9, шаг - 0.01
		//вес - 10, шаг - 0.00
		diffR = countDiff(0, settings.Weights.Rating, 150)
		plusR = 0.

		iteration = 0
		err       = ""
		filters   Animes
	)
	for iteration < 200 {

		filters = a.Copy().
			filterByDuration(minD-int(plusD), maxD+int(plusD)).
			filterByEpisodes(minE-int(plusE), maxE+int(plusE)).
			filterByYear(minY-int(plusY), maxY+int(plusY)).
			filterByRatings(minR-plusR, maxR+plusR).
			filterByGenres(settings.GenresArr()).
			filterByKind(settings.Kind).
			filterByOldRatings(settings.OldRating)

		if len(filters) != 0 {
			break
		}

		plusR += diffR
		plusD += diffD
		plusE += diffE
		plusY += diffY
		if iteration == 0 {
			err = "Нет тайтлов, удовлетворяющих выбранным условиям. Но мы подобрали небольшой список того, что может вам понравиться: "
		} else if iteration%50 == 0 {
			diffR *= 2
			diffD *= 2
			diffE *= 2
			diffY *= 2
		}
		iteration++
	}
	return filters, err
}

func (a Animes) filterByRatings(min, max float64) Animes {
	return a.filter(func(a Anime) bool {
		return a.Scoref >= min && a.Scoref <= max
	})
}

func (a Animes) filterByDuration(min, max int) Animes {
	return a.filter(func(a Anime) bool {
		return a.Duration >= min && a.Duration <= max
	})
}

func (a Animes) filterByEpisodes(min, max int) Animes {
	return a.filter(func(a Anime) bool {
		return a.Episodes >= min && a.Episodes <= max
	})
}

func (a Animes) filterByYear(min, max int) Animes {
	return a.filter(func(a Anime) bool {
		return a.Year >= min && a.Year <= max
	})
}

func (a Animes) filterByGenres(genres GenresMarked) Animes {
	return a.filter(func(a Anime) bool {
		for _, g1 := range genres {
			for _, g2 := range a.Genres {
				if g1.Marked && g1.Genre.ID == g2.ID {
					return true
				}
			}
		}
		return false
	})
}

func (a Animes) filterByOldRatings(oldRatings []IsMarked) Animes {
	return a.filter(func(a Anime) bool {
		for _, r1 := range oldRatings {
			if r1.Marked && r1.Value == a.Rating {
				return true
			}
		}
		return false
	})
}

func (a Animes) filterByKind(kinds []IsMarked) Animes {
	return a.filter(func(a Anime) bool {
		for _, r1 := range kinds {
			k := a.Kind
			if k == "tv_13" || k == "tv_24" || k == "tv_48" {
				k = "tv"
			}
			if r1.Marked && r1.Value == a.Kind {
				return true
			}
		}
		return false
	})
}

func (a Animes) filter(
	isOk func(a Anime) bool,
) Animes {
	var newAnimes = make(Animes, 0)
	for _, anime := range a {
		if isOk(anime) {
			newAnimes = append(newAnimes, anime)
		}
	}
	return newAnimes
}

func (a Animes) Len() int           { return len(a) }
func (a Animes) Less(i, j int) bool { return a[i].D < a[j].D }
func (a Animes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
