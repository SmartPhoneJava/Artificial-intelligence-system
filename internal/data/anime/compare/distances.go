package compare

import (
	"shiki/internal/data/anime"
)

type AnimeAllDistances struct {
	E, M, K, C, T, D AnimeDistances
}

func NewAnimeAllDistances(e, m, k, c, t, d AnimeDistances) AnimeAllDistances {
	return AnimeAllDistances{
		E: e,
		M: m,
		K: k,
		C: c,
		T: t,
		D: d,
	}
}

type AnimeDistance struct {
	D     float64
	Anime *anime.Anime
}

func NewDistance(d float64, a *anime.Anime) AnimeDistance {
	return AnimeDistance{
		D:     d,
		Anime: a,
	}
}

type AnimeDistances []AnimeDistance

func NewDistances(n int) AnimeDistances {
	return make([]AnimeDistance, n)
}

func (d AnimeDistances) Set(i int, f float64, a *anime.Anime) {
	d[i] = NewDistance(f, a)
}
func (d AnimeDistances) Len() int           { return len(d) }
func (d AnimeDistances) Less(i, j int) bool { return d[i].D < d[j].D }
func (d AnimeDistances) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
