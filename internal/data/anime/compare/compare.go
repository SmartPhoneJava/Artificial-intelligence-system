package compare

import (
	"shiki/internal/amath"
	"shiki/internal/data/anime"
	"shiki/internal/data/anime/settings"
	"sort"
)

type AnimeComparator struct {
	animes   anime.Animes
	settings settings.AnimeSettings

	pairs   []amath.Pairs
	pairFor int32
}

func NewAnimeComparator(a anime.Animes, s *settings.AnimeSettings) AnimeComparator {
	var sets settings.AnimeSettings
	if s == nil {
		sets = settings.NewAnimeSettings()
	} else {
		sets = *s
	}
	return AnimeComparator{
		animes:   a,
		settings: sets,
		tree:     tree,
	}
}

func (compare AnimeComparator) toPairs(first, second anime.Anime) amath.Pairs {
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

func (compare AnimeComparator) makePairs(first anime.Anime) {
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

func (compare AnimeComparator) distance(first anime.Anime, count func(p amath.Pairs) float64) AnimeDistances {
	compare.makePairs(first)
	var dists = NewDistances(len(compare.animes))
	for i, pair := range compare.pairs {
		dists.Set(i, count(pair), &compare.animes[i])
	}
	sort.Sort(dists)

	return dists
}

func (compare AnimeComparator) DistanceEuclideanAll(a anime.Anime) AnimeDistances {
	return compare.distance(a, func(p amath.Pairs) float64 { return p.Euclidean() })
}

func (compare AnimeComparator) DistanceL1All(a anime.Anime) AnimeDistances {
	return compare.distance(a, func(p amath.Pairs) float64 { return p.L1() })
}

func (compare AnimeComparator) DistanceChessAll(a anime.Anime) AnimeDistances {
	return compare.distance(a, func(p amath.Pairs) float64 { return p.Chess() })
}

func (compare AnimeComparator) DistanceDiffAll(a anime.Anime) AnimeDistances {
	return compare.distance(a, func(p amath.Pairs) float64 { return p.Diff() })
}

func (compare AnimeComparator) DistanceCorrelationAll(a anime.Anime) AnimeDistances {
	return compare.distance(a, func(p amath.Pairs) float64 { return p.Correlation() })
}

func (compare AnimeComparator) DistanceTreeAll(a anime.Anime) AnimeDistances {
	var dists = NewDistances(len(compare.animes))
	for i, anime := range compare.animes {
		dists.Set(i, float64(a.BranchDiff(anime)), &compare.animes[i])
	}
	sort.Sort(dists)
	return dists
}

func (compare AnimeComparator) DistanceAll(a anime.Anime) AnimeAllDistances {
	return NewAnimeAllDistances(
		compare.DistanceEuclideanAll(a),
		compare.DistanceL1All(a),
		compare.DistanceChessAll(a),
		compare.DistanceCorrelationAll(a),
		compare.DistanceTreeAll(a),
		compare.DistanceDiffAll(a),
	)
}
