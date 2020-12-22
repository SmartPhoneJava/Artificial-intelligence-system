package dialog

import (
	"errors"
	"log"
	"shiki/internal/anime"
	"shiki/internal/anime/recommend"
	"shiki/internal/models"
	"shiki/internal/page"
	"strconv"
)

const (
	// Информация о тайтле
	IntentDescription = "Узнать описание"
	IntentDuration    = "Узнать длительность серии"
	IntentYear        = "Узнать год выпуска"
	IntentOldRating   = "Узнать возрастной рейтинг"
	IntentScore       = "Узнать оценку"
	IntentStudio      = "Узнать студию производителя"
	IntentGenres      = "Узнать жанры"
	// Поставить оценку
	IntentRateTitle = "Оценка аниме"
	// Рекомендации
	IntentRecommendFilters = "Рекомендация по жанрам и тайтлам"
	IntentRecommendSimilar = "Рекомендация похожего"
	IntentRecommendMine    = "Рекомендация по моим предпочтениям"
)
const (
	EntityTitleName    = "title_name"
	EntityTitleContext = "context"
	EntityScore        = "score"
	EntityGenre        = "genre"
)

const (
	DefaultAnswer1 = "Извини, я тебя не понимаю:("
	DefaultAnswer2 = "Боюсь, тут я не могу помочь"
)

const (
	AnimeNotFound   = "Нямпасу, нам не удалось ничего найти :С"
	AnimeSayItAgain = "Напомни о каком аниме ты говоришь?"
	ScoreNotFound   = "Извини, не удалось понять понравился тебе тайтл или нет :С"
	AnimeRate10     = "Согласен, потрясное аниме!"
	AnimeRate9      = "Я рад, что это аниме тебе так понравилось"
	AnimeRate8      = "Круто, что тебе зашёл этот тайтл"
	AnimeRate7      = "Круто!"
	AnimeRate6      = "Я запомнил твои вкусы"
	AnimeRate5      = "Надеюсь следующее аниме тебе понравится больше"
	AnimeRate4      = "В следующий раз тебе повезет!"
	AnimeRate3      = "Жаль, что это аниме тебе не понравилось"
	AnimeRate2      = "Тоска..."
	AnimeRate1      = "Печально слышать это"
)

const MinConfidence = 0.5

func HandleResponseText(
	response NLPResponse,
	animes anime.AnimesUseCase,
	myScores models.UserScoreMap,
) (string, bool) {
	if response.Confidence < MinConfidence {
		return DefaultAnswer2, false
	}
	var msg string

	switch response.Intent {
	// Получить подробности про тайтл
	case IntentDescription:
		msg = getTitleInfo(response, animes, getDescription)
	case IntentDuration:
		msg = getTitleInfo(response, animes, getDuration)
	case IntentYear:
		msg = getTitleInfo(response, animes, getYear)
	case IntentOldRating:
		msg = getTitleInfo(response, animes, getOldRating)
	case IntentScore:
		msg = getTitleInfo(response, animes, getScore)
	case IntentGenres:
		msg = getTitleInfo(response, animes, getGenres)
	case IntentStudio:
		msg = getTitleInfo(response, animes, getStudio)
	// Оценить тайтл
	case IntentRateTitle:
		msg = setRate(response, animes, myScores)
		// Рекомендации
	case IntentRecommendMine, IntentRecommendFilters, IntentRecommendSimilar:
		return "", true
	default:
		msg = DefaultAnswer1
	}

	return msg, false
}

type HandledObject struct {
	Message string
	Animes  models.Animes
}

func HandleResponseObjects(
	response NLPResponse,
	input recommend.Input,
	genres []models.GenresMarked,
) (HandledObject, error) {
	var result HandledObject
	log.Println("Intent", response.Intent)
	switch response.Intent {
	// Получить подробности про тайтл
	case IntentRecommendMine:
		err := handleIntentRecommendMine(&result, input)
		return result, err
	case IntentRecommendSimilar:
		err := handleIntentRecommendSimilar(response, &result, input)
		return result, err
	case IntentRecommendFilters:
		err := handleIntentRecommendFilters(response, &result, input, genres)
		return result, err
	}
	return result, errors.New("No such intent type:" + response.Intent)
}

func handleIntentRecommendMine(
	result *HandledObject,
	input recommend.Input,
) error {
	var (
		recommendSettings = page.DefaultRecommendSettings
		recomendI         = recommend.NewCollaborativeFiltering(
			input,
			recommendSettings,
		)
	)
	animes, err := recomendI.Recommend()
	if err != nil {
		return err
	}
	//log.Println("IntentRecommendMine animes len", len(animes), len(myScores.Scores), len(allScores.Get()))
	result.Animes = animes
	result.Message = "Мне кажется тебе понравятся вот эти тайтлы"
	return nil
}

func handleIntentRecommendSimilar(
	response NLPResponse,
	result *HandledObject,
	input recommend.Input,
) error {
	foundAnime := nameToAnime(response, input.Animes)
	if foundAnime == nil {
		result.Message = AnimeSayItAgain
		return nil
	}
	var m = make(map[int]int)
	var s = models.UserScoreMap{
		Scores: m,
	}
	s.Add(int(foundAnime.ID), 10)

	var (
		weights  = models.DefaultWeigts()
		newInput = recommend.Input{
			Animes:   input.Animes,
			MyScores: s,
		}
		recomendI = recommend.NewContentOriented(
			newInput,
			&weights,
		)
	)
	animes, err := recomendI.Recommend()
	if err != nil {
		return err
	}
	if len(animes) == 0 {
		return handleIntentRecommendMine(result, input)
	}
	result.Animes = animes
	result.Message = "Мне кажется тебе понравятся вот эти тайтлы"

	return nil
}

func handleIntentRecommendFilters(
	response NLPResponse,
	result *HandledObject,
	input recommend.Input,
	genres []models.GenresMarked,
) error {
	var (
		findAnime = nameToAnime(response, input.Animes)
		newAnimes = input.Animes.Animes()
		err       error
	)
	log.Println("findAnime", findAnime)
	if findAnime != nil {
		var m = make(map[int]int)
		var s = models.UserScoreMap{
			Scores: m,
		}
		s.Add(int(findAnime.ID), 10)
		var (
			newInput = recommend.Input{
				Animes:   input.Animes,
				MyScores: s,
			}
			weights   = models.DefaultWeigts()
			recomendI = recommend.NewContentOriented(
				newInput,
				&weights,
			)
		)
		newAnimes, err = recomendI.Recommend()
		if err != nil {
			return err
		}
	}

	var (
		searchSettings = models.NewSimpleSearchSettings()
		genre          = response.Entities[EntityGenre]
	)
	searchSettings.Genres = make([]models.GenresMarked, len(genres))
	for i, g := range genres {
		searchSettings.Genres[i] = make(models.GenresMarked, len(g))
		for j, g := range g {
			searchSettings.Genres[i][j].Genre = g.Genre
			searchSettings.Genres[i][j].Marked = genre == g.Genre.Russian
		}
	}
	result.Animes, _ = newAnimes.Filter(searchSettings)
	result.Message = "Мне кажется тебе понравятся вот эти тайтлы"
	return err
}

// Изменить оценку тайтла
func setRate(
	response NLPResponse,
	animes anime.AnimesUseCase,
	myScores models.UserScoreMap,
) string {
	var (
		msg   string
		score = response.Entities[EntityScore]
		anime = nameToAnime(response, animes)
	)

	if anime == nil {
		msg = AnimeSayItAgain
	} else {
		scoreI, err := strconv.Atoi(score)
		if err != nil {
			msg = ScoreNotFound
		} else {
			myScores.Scores[int(anime.ID)] = scoreI
		}
		switch scoreI {
		case 10:
			msg = AnimeRate10
		case 9:
			msg = AnimeRate9
		case 8:
			msg = AnimeRate8
		case 7:
			msg = AnimeRate7
		case 6:
			msg = AnimeRate6
		case 5:
			msg = AnimeRate5
		case 4:
			msg = AnimeRate4
		case 3:
			msg = AnimeRate3
		case 2:
			msg = AnimeRate2
		case 1:
			msg = AnimeRate1
		}
	}
	return msg
}
