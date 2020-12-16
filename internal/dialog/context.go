package dialog

import (
	"shiki/internal/anime"
	"shiki/internal/models"
	"strings"
)

type Context struct {
	CurrentTitleName string
	CurrentAnime     *models.Anime
}

func (c *Context) SetName(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	c.CurrentTitleName = name
}

func (c Context) GetName() string {
	return c.CurrentTitleName
}

func (c Context) GetAnime(animes anime.AnimesUseCase) *models.Anime {
	if c.CurrentAnime != nil && c.CurrentAnime.SameName(c.CurrentTitleName) {
		return c.CurrentAnime
	}
	anime, found := animes.FindAnimeByName(c.CurrentTitleName)
	if !found {
		return nil
	}
	return &anime
}

var SingleContext = Context{}
