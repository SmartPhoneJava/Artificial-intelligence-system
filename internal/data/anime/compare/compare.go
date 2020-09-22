package compare

import (
	"shiki/internal/amath"
	"shiki/internal/data/anime"
	"shiki/internal/data/anime/settings"
)

type AnimeComparator struct {
	animes   anime.Animes
	settings settings.AnimeSettings
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
	pairs.AddSlice(first.Studios, second.Studios, settings.Studio)
	pairs.AddSlice(first.Genres, second.Genres, settings.Genre)
	return pairs
}

func (compare AnimeComparator) Euclidean(first, second anime.Anime) float64 {
	return compare.toPairs(first, second).Euclidean()
}
