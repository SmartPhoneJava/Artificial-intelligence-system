package compare

import (
	"errors"
	"log"
	"shiki/internal/models"
	"sort"

	"gonum.org/v1/gonum/floats"
)

// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.107.2790&rep=rep1&type=pdf
type CollaborativeFiltering struct {
	animes     models.Animes
	userScores models.UsersScoreMap
	myScore    models.UserScoreMap
}

func NewCollaborativeFiltering(
	animes models.Animes,
	userScores models.UsersScoreMap,
	myScore models.UserScoreMap,
) *CollaborativeFiltering {
	return &CollaborativeFiltering{
		animes:     animes,
		userScores: userScores,
		myScore:    myScore,
	}
}

func (filterring *CollaborativeFiltering) exceptMyAnimes() models.Animes {
	var newAnimes = make([]models.Anime, 0)
	for _, a := range filterring.animes {
		if filterring.myScore.Scores[int(a.ID)] == 0 {
			newAnimes = append(newAnimes, a)
		}
	}
	return newAnimes
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
			c += 1
		}
	}
	if n == 0 {
		return
	}

	anime.D = floats.Round(r/n, 2)
	anime.C = r
	anime.K = float64(c)
	log.Printf("-- %s - %.3f %.3f %.3f", anime.Name, anime.D, anime.C, anime.K)
}

func (filterring *CollaborativeFiltering) Recomend(
	usersCount int,
) (models.Animes, error) {

	if usersCount < 1 {
		return models.Animes{}, errors.New("usersCount < 1")
	}

	// компаратор для определения похожих юзеров
	comparator := NewUserComparator(
		filterring.userScores,
		filterring.myScore,
	)

	// отсортировали юзеров, по тому насколько их
	// вкусы совпадают с текущим юзером
	comparator.Sort(nil)

	var scores = make([]models.UserScoreMap, usersCount)
	copy(scores, comparator.peopleScores)

	// для нормализаци дистанции
	del := scores[len(scores)-1].D

	// тайтлы, которые юзер еще не видел
	filterring.animes = filterring.exceptMyAnimes()

	if len(filterring.animes) == 0 {
		return filterring.animes, nil
	}

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
