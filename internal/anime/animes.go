package anime

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/beevik/etree"

	"shiki/internal/anime/eatree"
	"shiki/internal/anime/shikimori"
	"shiki/internal/anime/tree"
	"shiki/internal/models"
	"shiki/internal/utils"
)

type AnimesUC struct {
	m   models.Animes
	a   []AnimeUseCase
	api shikimori.Api
}

func (auc *AnimesUC) init(m models.Animes) {
	var a = make([]AnimeUseCase, len(m))
	for i := range m {
		a[i] = NewAnime(&m[i])
	}
	auc.a = a
	auc.m = m
}

func NewAnimes(api shikimori.Api, m models.Animes) AnimesUseCase {
	var auc = new(AnimesUC)
	auc.init(m)
	auc.api = api
	return auc
}

func (auc AnimesUC) FetchDetails(
	ctx context.Context,
	UserAgent string,
	done chan error,
) {
	if len(auc.a) == 0 {
		err := auc.Load("")
		if err != nil {
			done <- err
			return
		}
	}
	for i := 0; i < len(auc.a); i++ {
		err := utils.MakeAction(ctx, func() error {
			return auc.a[i].FetchDetails(UserAgent)
		})
		log.Printf("Updated %d/%d", i, len(auc.a))
		if err != nil {
			done <- err
			return
		}
	}
	done <- nil
}

func (auc AnimesUC) Animes() models.Animes {
	return auc.m
}

func (auc AnimesUC) FindAnimes(
	name string,
) models.Animes {
	if name == "" {
		return auc.Animes()
	}
	var arr []models.Anime
	var found = make(map[int32]bool)
	for _, anime := range auc.m {
		if !found[anime.ID] && strings.Contains(anime.Russian, name) || strings.Contains(anime.Name, name) {
			arr = append(arr, anime)
			found[anime.ID] = true
		}
	}
	return arr
}

func (auc AnimesUC) FindAnimeByName(
	name string,
) (models.Anime, bool) {
	if name == "" {
		return models.Anime{}, false
	}
	for _, anime := range auc.m {
		if anime.Name == name || anime.Russian == name {
			return anime, true
		}
	}
	return models.Anime{}, false
}

func (auc AnimesUC) MarkMine(myScores models.UserScoreMap) {
	for i, v := range auc.m {
		score, ok := myScores.Scores[int(v.ID)]
		if ok {
			auc.m[i].IsMine = true
			auc.m[i].ScoreMine = score

			for j := 0; j < score; j++ {
				auc.m[i].Scorea[j] = struct {
					V bool
					I int
				}{true, j + 1}
			}

		} else {
			auc.m[i].IsMine = false
			for j := 0; j < 10; j++ {
				auc.m[i].Scorea[j] = struct {
					V bool
					I int
				}{false, j + 1}
			}
		}
	}
}

func (auc AnimesUC) UserAnimes(
	uc models.UserScoreMap,
) models.Animes {

	log.Println("len is", len(uc.Scores))

	var newAnimes = make([]models.Anime, 0)
	for _, anime := range auc.m {
		var score = uc.Scores[int(anime.ID)]
		if score > 0 {
			anime.Score = strconv.Itoa(score)
			newAnimes = append(newAnimes, anime)
		}
	}
	return newAnimes
}

func (auc AnimesUC) FindAnimeByID(
	id int32,
) (models.Anime, bool) {
	for _, anime := range auc.m {
		if anime.ID == id {
			return anime, true
		}
	}
	return models.Anime{}, false
}

func saveAnime(
	saveTo *etree.Element,
	anime models.Anime,
) error {
	if anime.Episodes == 0 {
		anime.Episodes = 1
	}
	bytesS, err := json.Marshal(anime)
	if err != nil {
		return err
	}
	saveTo.SetCData(string(bytesS))
	return nil
}

func (auc AnimesUC) Save(fromPath, toPath string) error {
	edoc, err := eatree.NewEdoc(fromPath)
	if err != nil {
		return err
	}
	for _, v := range edoc.Leaves {
		found, isFound := auc.FindAnimeByName(v.Name.Text())
		if isFound {
			err = saveAnime(v.Desription, found)
			if err != nil {
				return err
			}
		}
	}
	return edoc.Save(toPath)
}

func (auc *AnimesUC) Load(pathToFile string) error {
	var (
		path = "assets/res/cats_40.graphml"
	)
	if pathToFile != "" {
		path = pathToFile
	}
	edoc, err := eatree.NewEdoc(path)
	if err != nil {
		return err
	}

	var (
		added = make(map[int32]bool)
		m     models.Animes
	)

	for _, v := range edoc.Leaves {
		var anime models.Anime
		json.Unmarshal([]byte(v.Desription.Text()), &anime)
		anime.Branch = edoc.Tree.Branch(v.NodeID)
		anime.Scorea = make([]struct {
			V bool
			I int
		}, 10)

		for i := 0; i < 10; i++ {
			anime.Scorea[i] = struct {
				V bool
				I int
			}{false, i + 1}
		}

		if anime.ID > 0 && !added[anime.ID] {
			m = append(m, anime)
			added[anime.ID] = true
		}
	}
	auc.init(m)
	return nil
}

func createNode(
	graph *etree.Element,
	sourceID string,
	anime models.Anime,
) error {
	bytesArray, err := json.Marshal(anime)
	if err != nil {
		return err
	}
	var info = string(bytesArray)
	var nodes = graph.SelectElements("node")
	newID := "n" + strconv.FormatInt(int64(len(nodes)+1), 10)

	var newNode = nodes[0].Copy()
	newNode.Attr[0].Value = newID

	for _, v3 := range newNode.ChildElements() {
		for _, v4 := range v3.ChildElements() {
			for _, v5 := range v4.ChildElements() {
				if v5.Tag == "NodeLabel" {
					if anime.Russian != "" {
						v5.SetCData(anime.Russian)
					} else {
						v5.SetCData(anime.Name)
					}
				}
			}
		}
	}
	descr := newNode.CreateElement("data")
	descr.CreateAttr("key", "d5")
	descr.CreateAttr("xml:space", "preserve")
	descr.CreateCData(info)
	graph.AddChild(newNode)

	var edges = graph.SelectElements("edge")
	var newEdge = edges[0].Copy()
	newEdge.Attr[0].Value = "e" + utils.String(int32(len(edges)+1))
	newEdge.Attr[1].Value = sourceID
	newEdge.Attr[2].Value = newID
	graph.AddChild(newEdge)

	return nil
}

func (auc *AnimesUC) FetchData(
	ctx context.Context,
	fromPath, toPath string,
	limit int,
	tree tree.Tree,
	done chan<- error,
) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fromPath); err != nil {
		done <- err
		return
	}

	genres, err := models.NewGenres("internal/models/genres.json")
	if err != nil {
		done <- err
		return
	}

	studios, err := models.NewStudios("internal/models/studios.json")
	if err != nil {
		done <- err
		return
	}

	var animesFound = 0
	var animesShouldBe = limit * len(tree.Categories)
	for _, v := range doc.ChildElements() {
		for _, v1 := range v.ChildElements() {
			if v1.Tag == "graph" {
				for _, v2 := range v1.ChildElements() {
					if v2.Tag == "node" {
						name := tree.NodesNames[v2.Attr[0].Value]
						var category, ok = tree.Categories[name]
						if ok {
							err = utils.MakeAction(ctx, func() error {
								var (
									val    = v2.Attr[0].Value
									genre  string
									studio string
								)
								if category == "Жанры" {
									genre = utils.String(genres.ToID(name))
								} else if category == "Студия" {
									studio = utils.String(studios.ToID(name))

								}
								animes, err := auc.api.GetAnimes(int32(limit), genre, studio)
								if err != nil {
									return err
								}
								animesFound += len(animes)
								if len(animes) < 40 {
									animesShouldBe -= 40 - len(animes)
								}
								log.Printf("Received %d/%d",
									animesFound,
									animesShouldBe,
								)
								for _, anime := range animes {
									createNode(v1, val, anime)
								}
								return nil
							},
							)
							if err != nil {
								done <- err
								return
							}
						}
					}
				}
			}
		}
	}

	f, err := os.OpenFile(toPath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		done <- err
		return
	}
	_, err = doc.WriteTo(f)

	if err != nil {
		done <- err
		return
	}
	err = f.Close()
	if err != nil {
		done <- err
		return
	}
	done <- nil
}
