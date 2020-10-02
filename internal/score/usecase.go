package score

import (
	"context"
	"shiki/internal/models"
)

type UseCase interface {
	Load(path string) error
	Save(path string) error

	Fetch(
		ctx context.Context,
		users int32,
		done chan error,
	)

	Get() models.UsersScoreMap

	SetSettings(settings *models.ScoreSettings)
}
