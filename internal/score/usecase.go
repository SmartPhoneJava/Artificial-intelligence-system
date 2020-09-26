package score

import "shiki/internal/models"

type UseCase interface {
	Load(path string) error
	Save(path string) error

	Fetch(users, animes int32) error

	Get() models.UsersScoreMap
}
