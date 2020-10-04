package compare

import (
	"fmt"
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

	watched        []kv
	percentWatched float64
}

func NewUserComparator(
	peopleScores models.UsersScoreMap,
	myscores models.UserScoreMap,
	percentWatched float64,
) UserComparator {
	var uc = UserComparator{
		peopleScores:   peopleScores,
		myscores:       myscores,
		percentWatched: percentWatched,
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

func (uc UserComparator) NewPairs(mine, other []float64) amath.Pairs {
	pairs := make([]float64, 0, len(mine)*2)
	for i := 0; i < len(mine); i++ {
		if other[i] != 0 {
			pairs = pairs[:len(pairs)+2]
			pairs[len(pairs)-2] = mine[i]
			pairs[len(pairs)-1] = other[i]
		}
	}

	return pairs
}

func (uc UserComparator) floats(scores models.UserScoreMap) ([]float64, float64) {
	var (
		arr = make([]float64, len(uc.watched))
		i   = 0
		c   = len(uc.watched)
	)
	for _, v := range uc.watched {
		arr[i] = float64(scores.Scores[v.key])
		if arr[i] == 0 {
			c--
		}
		i++
	}
	return arr, float64(c) / float64(len(uc.watched))
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
		mine, _ = uc.floats(uc.myscores)
	)

	for i, v := range uc.peopleScores {
		another, count := uc.floats(v)

		if len(another) == 0 {
			uc.peopleScores[i].D = 10000
		} else if count < uc.percentWatched {
			uc.peopleScores[i].D = 5000 + 5000*(1-uc.percentWatched)
		} else {
			twoVectors := uc.NewPairs(mine, another)
			uc.peopleScores[i].D = (getDistance(twoVectors) + 1) * ((1 - count) + 0.001) * 1000
			if math.IsNaN(uc.peopleScores[i].D) {
				uc.peopleScores[i].D = 4000
			}
		}
	}
	sort.Sort(uc)
	for _, v := range uc.peopleScores {
		another, _ := uc.floats(v)
		fmt.Println("lllllll ", v.D, another)
	}
}

func (d *UserComparator) Len() int           { return len(d.peopleScores) }
func (d *UserComparator) Less(i, j int) bool { return d.peopleScores[i].D < d.peopleScores[j].D }
func (d *UserComparator) Swap(i, j int) {
	d.peopleScores[i], d.peopleScores[j] = d.peopleScores[j], d.peopleScores[i]
}
