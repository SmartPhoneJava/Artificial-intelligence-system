package router

import (
	"shiki/internal/anime/compare"

	"gonum.org/v1/gonum/floats"
)

func Mul(param1, param2 float64) float64 {
	return floats.Round(param1*param2, 3)
}

func Ec(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Ec
	}
}

func Mc(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Mc
	}
}

func Kc(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Kc
	}
}

func Dc(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Dc
	}
}

func Cc(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Cc
	}
}

func Tc(dists compare.AnimeAllDistances) func() bool {
	return func() bool {
		return dists.Tc
	}
}
