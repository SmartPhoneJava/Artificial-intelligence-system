package compare

import (
	"shiki/internal/amath"
	"shiki/internal/models"
	"sort"
)

type UserComparator struct {
	peopleScores models.UsersScoreMap
	myscores     models.UserScoreMap
}

func NewUserComparator(
	peopleScores models.UsersScoreMap,
	myscores models.UserScoreMap,
) UserComparator {
	var uc = UserComparator{
		peopleScores: peopleScores,
		myscores:     myscores,
	}
	return uc
}

func (uc UserComparator) scoresVector(scores models.UserScoreMap) []float64 {
	var (
		arr = make([]float64, len(uc.myscores.Scores))
		i   = 0
	)
	for k := range uc.myscores.Scores {
		arr[i] = float64(scores.Scores[k])
		i++
	}
	return arr
}

func (uc *UserComparator) Sort(
	getDistance func(amath.Pairs) float64,
) error {
	var (
		mine = uc.scoresVector(uc.myscores)
	)
	for i, v := range uc.peopleScores {
		another := uc.scoresVector(v)
		twoVectors, err := amath.NewPairs(mine, another)
		if err != nil {
			return err
		}
		uc.peopleScores[i].D = getDistance(twoVectors)
	}
	sort.Sort(uc.peopleScores)

	return nil
}
