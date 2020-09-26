package anime

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"shiki/internal/models"
	"shiki/internal/utils"
	"strconv"
	"time"
)

type AnimeUC struct {
	*models.Anime
}

func NewAnime(m *models.Anime) AnimeUseCase {
	return &AnimeUC{m}
}

func (anime AnimeUC) BranchDiff(another models.Anime) int {
	for i, t := range utils.InvertArr(anime.Branch) {
		for j, v := range utils.InvertArr(another.Branch) {
			if t == v {
				return i + j + 2
			}
		}
	}
	return len(anime.Branch) + len(another.Branch)
}

func (anime *AnimeUC) ratingToInt() int {
	switch anime.Rating {
	case "none":
		return 0
	case "g":
		return 2
	case "pg":
		return 8
	case "pg_13":
		return 13
	case "r":
		return 16
	case "r_plus":
		return 18
	case "rx":
		return 21
	}
	return 0
}

func (anime *AnimeUC) FetchDetails() error {
	client := &http.Client{}

	id := utils.String(anime.ID)
	req, _ := http.NewRequest(
		"GET",
		"https://shikimori.one/api/animes/"+id,
		nil,
	)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return utils.Err429
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, anime)
	if err != nil {
		return err
	}
	anime.Scoref, err = strconv.ParseFloat(anime.Score, 64)
	anime.RatingI = anime.ratingToInt()
	t, err := time.Parse("2006-01-02", anime.AiredOn)
	if err != nil {
		log.Println("time.Parse update error ", err)
		err = nil
		anime.Year = time.Now().Year()
	} else {
		anime.Year = t.Year()
	}
	log.Println("new anime is ", *anime.Anime)
	return err
}
