package recommend

import (
	"shiki/internal/anime/compare"
	"shiki/internal/models"
	"sort"

	"gonum.org/v1/gonum/floats"
)

type ContentOriented struct {
	animes  models.Animes
	myScore models.UserScoreMap

	compare compare.AnimeComparator
}

func NewContentOriented(
	animes models.Animes,
	myScore models.UserScoreMap,
	weights *models.Weigts,
) RecomendI {
	return &ContentOriented{
		animes:  animes,
		myScore: myScore,
		compare: compare.NewAnimeComparator(animes, weights),
	}
}

func (c *ContentOriented) Recommend() (models.Animes, error) {
	var m2m = c.processAllDistances()
	var idWithScore = c.oneAnimeManyDistances(m2m)
	c.updateD(idWithScore)
	return c.animes, nil
}

func (c *ContentOriented) processAllDistances() compare.ComparingAnimes {

	var m2m = make(compare.ComparingAnimes, 0, len(c.myScore.Scores))

	for id, score := range c.myScore.Scores {
		for _, anime := range c.animes {
			if int32(id) == anime.ID {
				dists := c.compare.DistanceEuclideanAll(anime)
				m2m = m2m[:len(m2m)+1]
				m2m[len(m2m)-1] = compare.ComparingAnime{
					ID:    id,
					Score: score,
					Dists: dists,
				}
				break
			}
		}
	}

	return m2m
}

func (c *ContentOriented) oneAnimeManyDistances(m2m compare.ComparingAnimes) map[int32]float64 {
	var animeDists = make(map[int32]float64, 0)
	for _, animes := range m2m {
		for _, danime := range animes.Dists {
			//log.Println("danime is", danime)
			animeDists[danime.Anime.ID] = danime.D * float64(11-animes.Score)
		}
	}

	for _, anime := range m2m {
		animeDists[int32(anime.ID)] = 1000000
	}
	return animeDists
}

func (c *ContentOriented) updateD(mapa map[int32]float64) {
	for i, anime := range c.animes {
		distance, ok := mapa[anime.ID]
		if ok {
			c.animes[i].D = distance
		}
	}
	sort.Sort(c)
	del := c.animes[len(c.animes)-1].D
	for i := range c.animes {

		c.animes[i].D = floats.Round(10-c.animes[i].D/del*10, 3)

	}
}

func (d *ContentOriented) Len() int { return len(d.animes) }
func (d *ContentOriented) Less(i, j int) bool {
	return d.animes[i].D < d.animes[j].D
}
func (d *ContentOriented) Swap(i, j int) { d.animes[i], d.animes[j] = d.animes[j], d.animes[i] }
