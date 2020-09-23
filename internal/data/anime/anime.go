package anime

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"shiki/internal/data/genres"
	"shiki/internal/data/studios"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
)

type Anime struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	Russian       string `json:"russian"`
	URL           string `json:"url"`
	Status        string `json:"status"`
	EpisodesAired int    `json:"episodes_aired"`
	AiredOn       string `json:"aired_on"`
	ReleasedOn    string `json:"released_on"`
	Score         string `json:"score"`
	Rating        string `json:"rating"` //!++

	Kind     string  `json:"kind"` //!++
	Scoref   float64 //!+
	Episodes int     `json:"episodes"` //!+
	Duration int     `json:"duration"` //!++
	RatingI  int     //!++
	Year     int     //!++
	Ongoing  bool    `json:"ongoing"`

	Studios studios.Studios `json:"studios"` //!+++
	Genres  genres.Genres   `json:"genres"`  //!+++
}

func (anime Anime) ToString() string {
	return fmt.Sprintf("\nНазвание:%s\nТип:%sРейтинг:%s\nКоличество эпизодов:%d\n",
		anime.Name,
		anime.Kind,
		anime.Score,
		anime.Episodes,
		anime.Duration,
		anime.Rating,
		anime.Year,
		anime.Ongoing,
		anime.Studios,
		anime.Genres,
	)
}

func (anime *Anime) ratingToInt() int {
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

func (anime *Anime) Update() error {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://shikimori.one/api/animes/"+String(anime.ID), nil)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	log.Println("resp.Status ", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Println("resp.body ", string(body))

	err = json.Unmarshal(body, anime)
	if err != nil {
		return err
	}
	anime.Scoref, err = strconv.ParseFloat(anime.Score, 64)
	anime.RatingI = anime.ratingToInt()
	t, err := time.Parse("2006-01-02", anime.AiredOn)
	if err != nil {
		return err
	}
	anime.Year = t.Year()
	return err
}

type Animes []Anime

func (animes Animes) Search(name string) []Anime {
	if name == "" {
		return animes
	}
	var arr []Anime
	for _, anime := range animes {
		if strings.Contains(anime.Russian, name) || strings.Contains(anime.Name, name) {
			arr = append(arr, anime)
		}
	}
	return arr
}

func (animes *Animes) Load(pathToFile string) error {
	var (
		path = "res/cats4.graphml"
		doc  = etree.NewDocument()
	)
	if pathToFile != "" {
		path = pathToFile
	}
	if err := doc.ReadFromFile(path); err != nil {
		return err
	}

	for _, v := range doc.ChildElements() {
		for _, v1 := range v.ChildElements() {
			if v1.Tag == "graph" {
				for _, v2 := range v1.ChildElements() {
					if v2.Tag == "node" {
						for _, v3 := range v2.ChildElements() {

							var flag bool
							for _, attr := range v3.Attr {

								if attr.Value == "d5" {
									flag = true

								}
							}

							if flag {
								var anime Anime
								json.Unmarshal([]byte(v3.Text()), &anime)
								*animes = append(*animes, anime)

							}

						}
					}
				}
			}
		}
	}
	return nil
}

func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
