package compare

import (
	"fmt"
	"shiki/internal/amath"
	"shiki/internal/anime"
	"shiki/internal/models"
)

type AnimeComparator struct {
	animes   models.Animes
	settings models.Weigts

	pairs   []amath.Pairs
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

func (compare AnimeComparator) toPairs(
	first, second models.Anime,
) amath.Pairs {
	var (
		pairs    amath.Pairs
		settings = compare.settings
	)
	pairs.AddString(first.Kind, second.Kind, settings.Kind)
	pairs.Add(first.Scoref, second.Scoref, settings.Score)
	pairs.AddInt(first.Episodes, second.Episodes, settings.Episodes)
	pairs.AddInt(first.Duration, second.Duration, settings.Duration)
	pairs.AddInt(first.RatingI, second.RatingI, settings.Rating)
	pairs.AddInt(first.Year, second.Year, settings.Year)
	pairs.AddBool(first.Ongoing, second.Ongoing, settings.Ongoing)
	pairs.AddSlice(first.Studios.Names(), second.Studios.Names(), settings.Studio)
	pairs.AddSlice(first.Genres.Names(), second.Genres.Names(), settings.Genre)

	return pairs
}

func (compare *AnimeComparator) makePairs(
	first models.Anime,
) {
	if compare.pairFor == first.ID {
		return
	}
	var pairs = make([]amath.Pairs, 0)
	for _, a := range compare.animes {
		if first.ID != a.ID {
			pairs = append(pairs, compare.toPairs(first, a))
		}
	}
	compare.pairs = pairs
	compare.pairFor = first.ID
}

func (compare AnimeComparator) distance(
	first models.Anime,
	count func(p amath.Pairs) float64,
) AnimeDistances {
	(&compare).makePairs(first)

	var dists = NewDistances(len(compare.animes))
	for i, pair := range compare.pairs {
		dists.Set(i, count(pair), &compare.animes[i])
	}
	return dists
}

func (compare AnimeComparator) DistanceEuclideanAll(
	a models.Anime,
) AnimeDistances {
	return compare.distance(a,
		func(p amath.Pairs) float64 { return p.Euclidean() },
	)
}

func (compare AnimeComparator) DistanceL1All(
	a models.Anime,
) AnimeDistances {
	return compare.distance(a,
		func(p amath.Pairs) float64 { return p.L1() },
	)
}

func (compare AnimeComparator) DistanceChessAll(
	a models.Anime,
) AnimeDistances {
	return compare.distance(a,
		func(p amath.Pairs) float64 { return p.Chess() },
	)
}

func (compare AnimeComparator) DistanceDiffAll(
	a models.Anime,
) AnimeDistances {
	return compare.distance(a,
		func(p amath.Pairs) float64 { return p.Diff() },
	)
}

func (compare AnimeComparator) DistanceCorrelationAll(
	a models.Anime,
) AnimeDistances {
	return compare.distance(a,
		func(p amath.Pairs) float64 { return p.Correlation() },
	)
}

func (compare AnimeComparator) DistanceTreeAll(
	a models.Anime,
) AnimeDistances {
	var dists = NewDistances(len(compare.animes))

	for i, v := range compare.animes {
		var dif = anime.NewAnime(&a).BranchDiff(v)
		dists.Set(i, float64(dif), &compare.animes[i])
	}
	return dists
}

func (compare AnimeComparator) DistanceAll(
	a models.Anime,
	ok bool,
) AnimeAllDistances {
	var (
		e = NewDistances(0)
		m = NewDistances(0)
		k = NewDistances(0)
		c = NewDistances(0)
		t = NewDistances(0)
		d = NewDistances(0)
	)

	if ok {
		fmt.Println("e")
		e = compare.DistanceEuclideanAll(a)
		m = compare.DistanceL1All(a)
		fmt.Println("k")
		k = compare.DistanceChessAll(a)
		c = compare.DistanceCorrelationAll(a)
		fmt.Println("c")
		t = compare.DistanceTreeAll(a)
		d = compare.DistanceDiffAll(a)
		fmt.Println("d")
	}
	return NewAnimeAllDistances(a, ok, e, m, k, c, t, d)
}
