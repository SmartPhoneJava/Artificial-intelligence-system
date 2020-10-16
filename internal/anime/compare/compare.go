package compare

import (
	"shiki/internal/amath"
	"shiki/internal/anime"
	"shiki/internal/models"

	"gonum.org/v1/gonum/floats"
)

type AnimeComparator struct {
	animes   models.Animes
	settings models.Weigts

	pairs   []ComparedAnime
	pairFor int32
}

func NewAnimeComparator(
	a models.Animes,
	s *models.Weigts,
) AnimeComparator {
	var sets models.Weigts
	if s == nil {
		sets = models.DefaultWeigts()
	} else {
		sets = *s
	}
	return AnimeComparator{
		animes:   a,
		settings: sets,
	}
}

func (compare AnimeComparator) toPairs(a, b models.Anime) amath.Pairs {
	var (
		pairs    amath.Pairs
		settings = compare.settings
	)
	pairs.AddString(a.Kind, b.Kind, settings.Kind)
	pairs.Add(a.Scoref, b.Scoref, settings.Score)
	pairs.AddInt(a.Episodes, b.Episodes, settings.Episodes)
	pairs.AddInt(a.Duration, b.Duration, settings.Duration)
	pairs.AddInt(a.RatingI, b.RatingI, settings.Rating)
	pairs.AddInt(a.Year, b.Year, settings.Year)
	pairs.AddBool(a.Ongoing, b.Ongoing, settings.Ongoing)
	pairs.AddSlice(a.Studios, b.Studios, amath.Linear, settings.Studio)
	pairs.AddSlice(a.Genres, b.Genres, amath.Square, settings.Genre)

	return pairs
}

func (compare *AnimeComparator) makePairs(
	first models.Anime,
) {
	if compare.pairFor == first.ID {
		return
	}
	var pairs = make([]ComparedAnime, 0)
	for i, a := range compare.animes {
		if first.ID != a.ID {
			pairs = append(pairs, ComparedAnime{
				pairs: compare.toPairs(first, a),
				anime: &compare.animes[i],
			})
		}
	}
	compare.pairs = pairs
	compare.pairFor = first.ID
}

// Возвращает массив схожести тайтлов с first
// first - тайтл, с которым происходит сравнение
// count - мера близости
func (compare *AnimeComparator) distance(
	first models.Anime,
	setter func(a *models.Anime, p amath.Pairs),
) {
	compare.makePairs(first)

	for _, v := range compare.pairs {
		setter(v.anime, v.pairs)
	}
}

func (compare *AnimeComparator) DistanceEuclideanAll(
	a models.Anime,
) *AnimeComparator {
	compare.distance(a,
		func(a *models.Anime, p amath.Pairs) {
			a.E = floats.Round(p.Euclidean(), 3)
		})
	return compare
}

func (compare *AnimeComparator) DistanceL1All(
	a models.Anime,
) *AnimeComparator {
	compare.distance(a,
		func(a *models.Anime, p amath.Pairs) {
			a.M = floats.Round(p.L1(), 3)
		})
	return compare
}

func (compare *AnimeComparator) DistanceChessAll(
	a models.Anime,
) *AnimeComparator {
	compare.distance(a,
		func(a *models.Anime, p amath.Pairs) {
			a.K = floats.Round(p.Chess(), 3)
		})
	return compare
}

func (compare *AnimeComparator) DistanceDiffAll(
	a models.Anime,
) *AnimeComparator {
	compare.distance(a,
		func(a *models.Anime, p amath.Pairs) {
			a.D = floats.Round(p.Diff(), 3)
		})
	return compare
}

func (compare *AnimeComparator) DistanceCorrelationAll(
	a models.Anime,
) *AnimeComparator {
	compare.distance(a,
		func(a *models.Anime, p amath.Pairs) {
			a.C = floats.Round(p.Correlation(), 3)
		})
	return compare
}

func (compare *AnimeComparator) DistanceTreeAll(
	a models.Anime,
) *AnimeComparator {
	for i, v := range compare.animes {
		var dif = anime.NewAnime(&a).BranchDiff(v)
		compare.animes[i].T = floats.Round(float64(dif), 3)
	}
	return compare
}

func (compare *AnimeComparator) DistanceAll(
	a models.Anime,
) *AnimeComparator {
	return compare.
		DistanceEuclideanAll(a).
		DistanceL1All(a).
		DistanceCorrelationAll(a).
		DistanceChessAll(a).
		DistanceTreeAll(a).
		DistanceDiffAll(a)
}

func (compare AnimeComparator) Animes() models.Animes {
	return compare.animes
}
