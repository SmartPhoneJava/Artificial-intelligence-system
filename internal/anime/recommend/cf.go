package recommend

import (
	"errors"
	"fmt"
	"shiki/internal/anime/compare"
	"shiki/internal/models"
	"shiki/internal/page"
	"sort"

	"gonum.org/v1/gonum/floats"
)

// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.107.2790&rep=rep1&type=pdf
type CollaborativeFiltering struct {
	animes     models.Animes
	userScores models.UsersScoreMap
	myScore    models.UserScoreMap

	settings page.RecommendSettings
}

func NewCollaborativeFiltering(
	input Input,
	settings page.RecommendSettings,
) RecomendI {
	return &CollaborativeFiltering{
		animes:     input.Animes.Animes(),
		userScores: input.AllScores.Get(),
		myScore:    input.MyScores,
		settings:   settings,
	}
}

func (filterring *CollaborativeFiltering) aggregation(
	anime *models.Anime,
	scores []models.UserScoreMap,
	del float64,
) {
	var (
		r float64
		n float64
		c int
	)
	if del == 0 {
		del = 1
	}

	for _, cs := range scores {
		score := cs.Scores[int(anime.ID)]
		if score > 0 {
			var (
				weight = 1 - (cs.D/del)*4/5
			)

			r += weight * float64(score)
			n += weight
			c++
		}
	}
	if n == 0 {
		return
	}

	anime.D = floats.Round(r/n, 2)
	anime.C = r
	anime.K = float64(c)
}

func (filterring *CollaborativeFiltering) Recommend() (models.Animes, error) {

	if filterring.settings.Users < 1 {
		return models.Animes{}, errors.New("usersCount < 1")
	}

	// компаратор для определения похожих юзеров
	// отсортировали юзеров, по тому насколько их
	// вкусы совпадают с текущим юзером
	peopleScores := compare.NewUserComparator(
		filterring.userScores,
		filterring.myScore,
		filterring.settings.Percent,
	).Sort(nil)

	// уберем лишних юзеров
	var scores = make([]models.UserScoreMap, filterring.settings.Users)
	copy(scores, peopleScores)

	// для нормализаци дистанции
	del := scores[len(scores)-1].D

	// тайтлы, которые юзер еще не видел
	filterring.animes = filterring.myScore.ExceptMine(filterring.animes)

	if len(filterring.animes) == 0 {
		return filterring.animes, nil
	}

	fmt.Println("del is", del)
	for i := range filterring.animes {
		filterring.aggregation(&filterring.animes[i], scores, del)
	}
	sort.Sort(filterring)

	return filterring.animes, nil
}

func (d *CollaborativeFiltering) Len() int { return len(d.animes) }
func (d *CollaborativeFiltering) Less(i, j int) bool {
	return d.animes[i].D > d.animes[j].D ||
		(d.animes[i].D == d.animes[j].D && d.animes[i].C > d.animes[j].C)
}
func (d *CollaborativeFiltering) Swap(i, j int) { d.animes[i], d.animes[j] = d.animes[j], d.animes[i] }
