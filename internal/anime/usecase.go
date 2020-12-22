package anime

import (
	"context"
	"shiki/internal/anime/tree"
	"shiki/internal/models"
)

type AnimeUseCase interface {
	BranchDiff(another models.Anime) int

	FetchDetails(UserAgent string) error
}

type AnimesUseCase interface {
	FetchData(
		ctx context.Context,
		fromPath, toPath string,
		limit int,
		tree tree.Tree,
		done chan<- error,
	)
	FetchDetails(ctx context.Context, UserAgent string, done chan error)

	Animes() models.Animes
	FilterByName(name string) models.Animes
	FindOneAnime(name string) (models.Anime, bool)
	FindAnimeByName(name string) (models.Anime, bool)
	FindAnimeByID(id int32) (models.Anime, bool)

	MarkMine(myScores models.UserScoreMap)

	UserAnimes(uc models.UserScoreMap) models.Animes

	Save(fromPath, toPath string) error
	Load(pathToFile string) error
}
