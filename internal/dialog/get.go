package dialog

import (
	"fmt"
	"shiki/internal/anime"
	"shiki/internal/models"
)

// Получить аниме на основе запроса и контекста
func nameToAnime(
	response NLPResponse,
	animes anime.AnimesUseCase,
) *models.Anime {
	var (
		titleNames = response.Entities[EntityTitleName]
		titleName  string
	)

	if len(titleNames) != 0 {
		titleName = titleNames[0]
	}

	SingleContext.SetName(titleName)
	titleName = SingleContext.GetName()
	return SingleContext.GetAnime(animes)
}

func namesToAnime(
	response NLPResponse,
	animesUC anime.AnimesUseCase,
) models.Animes {
	var (
		titleNames = response.Entities[EntityTitleName]
		animes     = make(models.Animes, 0)
	)
	for _, titleName := range titleNames {
		anime, found := animesUC.FindOneAnime(titleName)
		if found {
			animes = append(animes, anime)
		}
	}

	return animes
}

// Получить информацию о тайтле
func getTitleInfo(
	response NLPResponse,
	animes anime.AnimesUseCase,
	get func(models.Anime) string,
) string {
	var msg string
	anime := nameToAnime(response, animes)
	if anime == nil {
		msg = AnimeNotFound
	} else {
		msg = get(*anime)
	}
	return msg
}

// anime get info funcs

func getDescription(anime models.Anime) string {
	return "Описание аниме(" + anime.Russian + "):" + anime.DescriptionHTML
}

func getDuration(anime models.Anime) string {
	return fmt.Sprintf("Длительность аниме(%s): %d серий %d минут",
		anime.Russian, anime.Episodes, anime.Duration)
}

func getYear(anime models.Anime) string {
	return fmt.Sprintf("'%s' вышло в %d", anime.Russian, anime.Year)
}

func getOldRating(anime models.Anime) string {
	return fmt.Sprintf("Возрастная оценка '%s': %s", anime.Russian, anime.Rating)
}

func getScore(anime models.Anime) string {
	return fmt.Sprintf("%s имеет оценку %s/10 баллов на shikimori", anime.Russian, anime.Score)
}

func getStudio(anime models.Anime) string {
	names := anime.Studios.Names()
	if len(names) == 0 {
		return fmt.Sprintf("Аниме '%s' выпущено безымянной студией", anime.Russian)
	}
	if len(names) == 1 {
		return fmt.Sprintf("Аниме '%s' выпущено студией %s", anime.Russian, names[0])
	}
	msg := fmt.Sprintf("Аниме '%s' выпущено следующими студиями:", anime.Russian)
	for _, name := range names {
		msg += " " + name
	}
	return msg
}

func getGenres(anime models.Anime) string {
	names := anime.Genres.Names()
	if len(names) == 0 {
		return fmt.Sprintf("Аниме '%s' нельзя отнести ни к одному из существующих на сегодняшний день жанров", anime.Russian)
	}

	msg := fmt.Sprintf("Аниме '%s' относится к жанрам:", anime.Russian)
	for _, name := range names {
		msg += " " + name
	}
	return msg
}
