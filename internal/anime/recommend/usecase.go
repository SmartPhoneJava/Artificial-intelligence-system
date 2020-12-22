package recommend

import (
	"shiki/internal/anime"
	"shiki/internal/models"
	"shiki/internal/score"
)

type Input struct {
	Animes    anime.AnimesUseCase
	MyScores  models.UserScoreMap
	AllScores score.UseCase
}

type RecomendI interface {
	Recommend() (models.Animes, error)
}
