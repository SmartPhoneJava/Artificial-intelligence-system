package anime

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"shiki/internal/data/anime/eatree"
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
	Branch  []string        `json:"branch"`  //!+++
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

var Err429 = errors.New("429 Too Many Requests")

func (anime *Anime) Update() error {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://shikimori.one/api/animes/"+String(anime.ID), nil)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	log.Println("resp.Status ", resp.Status)
	if resp.StatusCode == http.StatusTooManyRequests {
		return Err429
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
		return err
	}
	anime.Year = t.Year()
	log.Println("resp.body ", anime)
	return err
}

type Animes []Anime

func (animes Animes) Update(done chan struct{}) {
	log.Printf("we begin update", len(animes))
	var count = 0
	i := 0
	if len(animes) == 0 {
		animes.Load("")
	}
	for i < len(animes) {
		log.Printf("%d/%d", i, len(animes))
		if count > 0 && count%90 == 0 {
			time.Sleep(time.Second * 50)
		}
		count++
		time.Sleep(100 * time.Millisecond)
		err := (&animes[i]).Update()
		if err == Err429 {

		} else {
			i++
		}
	}
	log.Printf("%d/%d", i, len(animes))
	done <- struct{}{}
	log.Printf("we end update", len(animes))
}

func (animes Animes) SearchByName(name string) Animes {
	if name == "" {
		return animes
	}
	var arr []Anime
	var found = make(map[int32]bool)
	log.Println("search!", name)
	for i, anime := range animes {

		log.Println("www! ", i, strings.Contains(anime.Name, name), anime.Name, name)
		if !found[anime.ID] && strings.Contains(anime.Russian, name) || strings.Contains(anime.Name, name) {
			log.Println("contain!", anime.Name, name)
			arr = append(arr, anime)
			found[anime.ID] = true
		}
	}
	return arr
}

func (animes Animes) FindAnime(name string) (Anime, bool) {

	if name == "" {
		return Anime{}, false
	}
	for _, anime := range animes {
		if anime.Name == name || anime.Russian == name {
			return anime, true
		}
	}
	return Anime{}, false
}

func saveAnime(saveTo *etree.Element, anime Anime) error {
	anime.Duration = 777
	bytesS, err := json.Marshal(anime)
	if err != nil {
		return err
	}
	saveTo.SetCData(string(bytesS))
	return nil
}

func (animes Animes) Save(fromPath, toPath string) error {
	edoc, err := eatree.NewEdoc(fromPath)
	if err != nil {
		return err
	}
	for _, v := range edoc.Leaves {
		found, isFound := animes.FindAnime(v.Name.Text())
		if isFound {
			err = saveAnime(v.Desription, found)
			if err != nil {
				return err
			}
		}
	}
	log.Println("SAVE END")
	return edoc.Save(toPath)
}

func (animes *Animes) Load(pathToFile string) error {
	var (
		path = "assets/res/cats_3.graphml"
	)
	if pathToFile != "" {
		path = pathToFile
	}
	edoc, err := eatree.NewEdoc(path)
	if err != nil {
		return err
	}

	var added = make(map[int32]bool)

	for _, v := range edoc.Leaves {
		var anime Anime
		json.Unmarshal([]byte(v.Desription.Text()), &anime)
		log.Println("anime is", anime)
		anime.Branch = edoc.Tree.Branch(v.NodeID)
		if anime.ID > 0 && !added[anime.ID] {
			*animes = append(*animes, anime)
			added[anime.ID] = true
		}
	}

	return nil
}

// func (animes *Animes) LoadOld(pathToFile string) error {
// 	var (
// 		path = "res/cats4.graphml"
// 		doc  = etree.NewDocument()
// 	)
// 	if pathToFile != "" {
// 		path = pathToFile
// 	}
// 	if err := doc.ReadFromFile(path); err != nil {
// 		return err
// 	}

// 	var graphml = new(graphml.Graphml)
// 	err := graphml.Load(pathToFile)
// 	if err != nil {
// 		return err
// 	}

// 	var t = tree.NewTree()
// 	t.FromGraphml(*graphml, &tree.TreeSettings{
// 		LeavesKnown: true,
// 	})

// 	for _, v := range doc.ChildElements() {
// 		for _, v1 := range v.ChildElements() {
// 			if v1.Tag == "graph" {
// 				for _, v2 := range v1.ChildElements() {
// 					if v2.Tag == "node" {
// 						for _, v3 := range v2.ChildElements() {
// 							var flag bool
// 							for _, attr := range v3.Attr {
// 								if attr.Value == "d5" {
// 									flag = true
// 								}
// 							}

// 							if flag {
// 								var anime Anime

// 								json.Unmarshal([]byte(v3.Text()), &anime)
// 								if anime.ID > 0 {
// 									*animes = append(*animes, anime)
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

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
