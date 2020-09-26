package compare

import (
	"sort"
	"strings"

	"shiki/internal/models"
)

type AnimeAllDistances struct {
	Anime                  models.Anime
	E, M, K, C, T, D       AnimeDistances
	OK                     bool
	Ec, Mc, Kc, Cc, Tc, Dc bool
	animes                 []models.Anime
}

func (d AnimeAllDistances) Len() int { return len(d.animes) }
func (d AnimeAllDistances) Less(i, j int) bool {
	// обычные меры близости (proximity measures)
	var pm1, pm2 float64
	if d.Ec {
		pm1 += d.animes[i].E
		pm2 += d.animes[j].E
	}
	if d.Mc {
		pm1 += d.animes[i].M
		pm2 += d.animes[j].M
	}
	if d.Kc {
		pm1 += d.animes[i].K
		pm2 += d.animes[j].K
	}
	if d.Dc {
		pm1 += d.animes[i].D
		pm2 += d.animes[j].D
	}
	if d.Cc {
		pm1 += d.animes[i].C
		pm2 += d.animes[j].C
	}
	// Сравнение на основе близости по дереву самое приоритетное
	var t1, t2 float64
	if d.Tc {
		t1 = d.animes[i].T
		t2 = d.animes[j].T

		if t1 != t2 {
			return t1 < t2
		}
	}
	return pm1 < pm2
}
func (d AnimeAllDistances) Swap(i, j int) {
	d.animes[i], d.animes[j] = d.animes[j], d.animes[i]
}

func (aad AnimeAllDistances) Animes() models.Animes {
	var arr models.Animes
	for i := 0; i < len(aad.E); i++ {
		var comparingAnime = aad.E[i].Anime
		if comparingAnime != nil && comparingAnime.ID != aad.Anime.ID {
			comparingAnime.WithСoefficients(aad.E[i].D, aad.M[i].D, aad.K[i].D, aad.C[i].D, aad.T[i].D, aad.D[i].D)
			arr = append(arr, *comparingAnime)
		}
	}
	aad.animes = arr
	sort.Sort(aad)
	return aad.animes
}

func (aad *AnimeAllDistances) SetFilter(filter string) {
	aad.Ec = strings.Contains(filter, "e")
	aad.Mc = strings.Contains(filter, "m")
	aad.Kc = strings.Contains(filter, "k")
	aad.Cc = strings.Contains(filter, "c")
	aad.Tc = strings.Contains(filter, "t")
	aad.Dc = strings.Contains(filter, "d")
}

func NewAnimeAllDistances(
	a models.Anime,
	ok bool,
	e, m, k, c, t, d AnimeDistances,
) AnimeAllDistances {
	return AnimeAllDistances{
		E: e, Ec: true,
		M: m, Mc: false,
		K: k, Kc: false,
		C: c, Cc: false,
		T: t, Tc: false,
		D: d, Dc: false,
		Anime: a, OK: ok,
	}
}

type AnimeDistance struct {
	D     float64
	Anime *models.Anime
}

func NewDistance(d float64, a *models.Anime) AnimeDistance {
	return AnimeDistance{
		D:     d,
		Anime: a,
	}
}

type AnimeDistances []AnimeDistance

func NewDistances(n int) AnimeDistances {
	return make([]AnimeDistance, n)
}

func (d AnimeDistances) Set(i int, f float64, a *models.Anime) {
	d[i] = NewDistance(f, a)
}
func (d AnimeDistances) Len() int { return len(d) }
func (d AnimeDistances) Less(i, j int) bool {
	return d[i].D < d[j].D
}
func (d AnimeDistances) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
