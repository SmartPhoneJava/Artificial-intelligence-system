package compare

import (
	"math"
	"shiki/internal/amath"
	"shiki/internal/models"
	"sort"
)

type kv struct {
	key, value int
}

type UserComparator struct {
	peopleScores models.UsersScoreMap
	myscores     models.UserScoreMap

	watched []kv
}

func NewUserComparator(
	peopleScores models.UsersScoreMap,
	myscores models.UserScoreMap,
) UserComparator {
	var uc = UserComparator{
		peopleScores: peopleScores,
		myscores:     myscores,
	}
	var (
		arr = make([]kv, len(uc.myscores.Scores))
		i   = 0
	)
	for k, v := range uc.myscores.Scores {
		arr[i] = kv{k, v}
		i++
	}
	uc.watched = arr
	return uc
}

func (uc UserComparator) floats(scores models.UserScoreMap) []float64 {
	var (
		arr = make([]float64, len(uc.watched))
		i   = 0
	)
	for _, v := range uc.watched {
		arr[i] = float64(scores.Scores[v.key])
		i++
	}
	return arr
}

func (uc *UserComparator) Sort(
	getDistance func(amath.Pairs) float64,
) {
	if getDistance == nil {
		getDistance = func(twoVectore amath.Pairs) float64 {
			if len(twoVectore) < 20 {
				return twoVectore.Euclidean()
			} else {
				return twoVectore.Correlation()
			}
		}
	}
	var (
		mine = uc.floats(uc.myscores)
	)

	for i, v := range uc.peopleScores {
		another := uc.floats(v)

		if len(another) == 0 {
			uc.peopleScores[i].D = 1000
		} else {
			twoVectors := amath.NewPairs(mine, another)
			uc.peopleScores[i].D = getDistance(twoVectors)
			if math.IsNaN(uc.peopleScores[i].D) {
				uc.peopleScores[i].D = 800
			}
		}
	}
	sort.Sort(uc)
}

func (d *UserComparator) Len() int           { return len(d.peopleScores) }
func (d *UserComparator) Less(i, j int) bool { return d.peopleScores[i].D < d.peopleScores[j].D }
func (d *UserComparator) Swap(i, j int) {
	d.peopleScores[i], d.peopleScores[j] = d.peopleScores[j], d.peopleScores[i]
}
