package compare

import (
	"errors"
	"shiki/internal/amath"
	"shiki/internal/models"
	"sort"
)

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

func (filterring *CollaborativeFiltering) removeMyAnimes() models.Animes {
	var newAnimes = make([]models.Anime, 0)
	for _, a := range filterring.animes {
		if filterring.myScore.Scores[int(a.ID)] == 0 {
			newAnimes = append(newAnimes, a)
		}
	}
	return newAnimes
}

func (filterring *CollaborativeFiltering) Recomend(
	usersCount int,
) (models.Animes, error) {
	newAnimes := filterring.removeMyAnimes()
	if len(newAnimes) == 0 {
		return newAnimes, nil
	}
	if usersCount < 1 {
		return newAnimes, errors.New("usersCount < 1")
	}

	comparator := NewUserComparator(
		filterring.userScores,
		filterring.myScore,
	)
	err := comparator.Sort(
		func(twoVectore amath.Pairs) float64 {
			return twoVectore.Euclidean()
		},
	)
	if err != nil {
		return newAnimes, err
	}
	animesCopy := make(models.Animes, len(newAnimes))
	var scores []models.UserScoreMap
	if usersCount >= len(comparator.peopleScores) {
		scores = comparator.peopleScores
	} else {
		scores = comparator.peopleScores[:usersCount]
	}

	// для нормализаци дистанции
	del := scores[len(scores)-1].D

	copy(animesCopy, newAnimes)
	for i, a := range animesCopy {
		var (
			summScore float64
			count     float64
		)

		for _, s := range scores {
			if s.Scores[int(a.ID)] > 0 {
				var (
					userK = 1 - s.D/del
					score = userK * float64(s.Scores[int(a.ID)])
				)
				summScore += score
				count++
			}
		}
		if count > 0 {
			animesCopy[i].D = summScore / count
			animesCopy[i].C = summScore
			animesCopy[i].K = count
		}
	}
	sort.Sort(animesCopy)
	return animesCopy, nil
}
