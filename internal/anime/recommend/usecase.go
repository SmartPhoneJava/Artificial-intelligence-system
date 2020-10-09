package recommend

import (
	"shiki/internal/models"
)

type RecomendI interface {
	Recommend() (models.Animes, error)
}
