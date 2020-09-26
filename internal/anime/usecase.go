package anime

import (
	"context"
	"shiki/internal/anime/tree"
	"shiki/internal/models"
)

type AnimeUseCase interface {
	BranchDiff(another models.Anime) int

	FetchDetails() error
}

type AnimesUseCase interface {
	FetchData(
		ctx context.Context,
		fromPath, toPath string,
		limit int,
		tree tree.Tree,
		done chan<- error,
	)
	FetchDetails(ctx context.Context, done chan error)

	Animes() models.Animes
	FindAnimes(name string) models.Animes
	FindAnimeByName(name string) (models.Anime, bool)
	FindAnimeByID(id int32) (models.Anime, bool)

	Save(fromPath, toPath string) error
	Load(pathToFile string) error
}
