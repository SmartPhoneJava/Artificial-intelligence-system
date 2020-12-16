package recommend

import (
	"log"
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

	// Получаем матрицу похожести A[i,j],
	// строка - тайтл из просмотренного,
	// столбец - тайтл, с которым сравниваем
	// в ячейке мера близости между этими тайтлами
	var m2m = c.processAllDistances()

	// Для каждого тайтла(столбца) получаем средневзвешенное меру близости
	// Результатом будет мэпа размером как у c.animes
	// с ключом: ID тайтла, значением: мерой
	var idWithScore = c.oneAnimeManyDistances(m2m)

	// Нормируем меру близости от 0 до 10
	// Записываем ее в поле сортировки .D и сортируем
	c.updateD(idWithScore)

	// убираем просмотренные тайтлы
	c.animes = c.myScore.ExceptMine(c.animes)

	return c.animes, nil
}

func (c *ContentOriented) processAllDistances() compare.ComparingAnimes {

	var m2m = make(compare.ComparingAnimes, 0, len(c.myScore.Scores))

	for id, score := range c.myScore.Scores {
		for _, anime := range c.animes {
			if int32(id) == anime.ID {
				dists := c.compare.DistanceEuclideanAll(anime).Animes()
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
			animeDists[danime.ID] = (danime.D + 1) * float64(11-animes.Score)
		}
	}
	return animeDists
}

func (c *ContentOriented) updateD(animeDists map[int32]float64) {

	for i, anime := range c.animes {
		distance, ok := animeDists[anime.ID]
		if ok {
			c.animes[i].D = distance
			log.Println("c.animes[i].D", distance)
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
