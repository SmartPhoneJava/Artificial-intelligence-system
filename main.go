package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"shiki/internal/data/anime"
	"shiki/internal/data/anime/compare"
	"shiki/internal/data/anime/tree"
	"shiki/internal/data/genres"
	"shiki/internal/data/studios"
	"shiki/internal/graphml"
	"shiki/internal/page"
	"strconv"
	"time"

	"github.com/beevik/etree"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
)

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectURL  string `json:"redirect_uri"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	Tokentype    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

const (
	UserAgent    = "Shikimori chrome extension"
	ClientID     = "5427ddff1021ee49d222b276e3f27fc0743896579dc8a79af2c76cfa94781e34"
	ClientSecret = "3ac12dfd95d24dba4ed581f9e47f0ea84bc4d9ed64a9c486be9de0cd0b55b726"

	Addr         = ":2997"
	ReadTimeout  = time.Second * 60
	WriteTimeout = time.Second * 60
	IdleTimeout  = time.Second * 60
)

var Token = TokenResponse{
	AccessToken:  "l5xR_Nl-tt4FdT_WX5sxLXnSCQ21B7JPQbI_QjRAYrw",
	Tokentype:    "Bearer",
	ExpiresIn:    86400,
	RefreshToken: "RyBPMUV5g-JwYkEJfL16l1ppxYeRI8CHKnZNShTYP68",
	Scope:        "user_rates comments topics",
	CreatedAt:    1600339489,
}

func findAnime(category, value string, limit int32, g genres.Genres, s studios.Studios) (anime.Animes, error) {
	client := &http.Client{}
	var query = "https://shikimori.one/api/animes?page=1&limit=" + String(limit) + "&censored=false&"
	if category == "Жанры" {
		query += "genre=" + String(g.ToID(value))
	}
	if category == "Студия" {
		query += "studio=" + String(s.ToID(value))
	}

	req, _ := http.NewRequest("GET", query, nil)
	// req.Header.Add("Accept", "application/json")
	// req.Header.Set("User-Agent", UserAgent)
	// req.Header.Set("Authorization", "Bearer "+Token.AccessToken)

	var animes anime.Animes

	resp, err := client.Do(req)
	if err != nil {
		return animes, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return animes, anime.Err429
	} else if resp.StatusCode != http.StatusOK {
		return animes, errors.New("Wrong status:" + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return animes, err
	}

	err = json.Unmarshal(body, &animes)
	if err != nil {
		return animes, err
	}
	return animes, nil
}

func createNode(graph *etree.Element, sourceID string, anime anime.Anime) error {
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
	newEdge.Attr[0].Value = "e" + strconv.FormatInt(int64(len(edges)+1), 10)
	newEdge.Attr[1].Value = sourceID
	newEdge.Attr[2].Value = newID
	graph.AddChild(newEdge)

	return nil
}

func makeAction(
	ctx context.Context,
	v1, v2 *etree.Element,
	category, name string,
	limit, animesShouldBe int,
	animesFound *int,
	genres genres.Genres, studios studios.Studios,
) error {

	var count = 0
	for {
		var timer = time.Millisecond * 100
		if count > 5 {
			count = 0
			timer = time.Minute
			log.Printf("RPM limit: Minute sleep")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timer):
			val := v2.Attr[0].Value
			animes, err := findAnime(category, name, int32(limit), genres, studios)
			if err == anime.Err429 {
				count++
				continue
			}
			count = 0
			*animesFound += len(animes)
			log.Printf("Received %d/%d", *animesFound, animesShouldBe)
			if err != nil {
				return err
			}
			for _, anime := range animes {
				createNode(v1, val, anime)
			}
			return nil
		}
	}

}

func change(
	ctx context.Context,
	fromPath, toPath string,
	tree tree.Tree,
	done chan<- error,
) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fromPath); err != nil {
		done <- err
		return
	}

	genres, err := genres.NewGenres("internal/data/genres/genres.json")
	if err != nil {
		done <- err
		return
	}

	studios, err := studios.NewStudios("internal/data/studios/studios.json")
	if err != nil {
		done <- err
		return
	}

	var limit int = 3
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
							err = makeAction(
								ctx,
								v1, v2,
								category, name,
								limit, animesShouldBe,
								&animesFound,
								genres, studios)
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

var requests = 0

func main() {
	router()
}

var ANIMES anime.Animes

func router() {
	r := mux.NewRouter()

	var tpl, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = ANIMES.Load("")
	if err != nil {
		log.Println("err is", err)
	}
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		animesFound := ANIMES.SearchByName(page.Settings.Search)
		tpl.Execute(w, struct {
			Animes anime.Animes
			Page   page.PageSettings
		}{
			Animes: animesFound,
			Page:   page.Settings,
		})
	})
	r.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.URL.Query() {
			switch k {
			case "tag":
				page.Settings.Tag = v[0]
				break
			case "search":
				page.Settings.Search = v[0]
				break
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	r.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		go ANIMES.Update(done)
		<-done
		ANIMES.Save("res/cats_3.graphml", "res/cats_3.graphml")
		log.Println("FINISH")
		//w.WriteHeader(http.StatusOK)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/token/{token}", func(w http.ResponseWriter, r *http.Request) {

		jsonStr, err := json.Marshal(TokenRequest{
			GrantType:    "authorization_code",
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			Code:         chi.URLParam(r, "token"),
			RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		})
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest("POST", "https://shikimori.one/oauth/token", bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, &Token)
		if err != nil {
			log.Fatal(err)
		}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "token", Value: string(body), Expires: expiration}
		http.SetCookie(w, &cookie)
		w.Write([]byte("Вы авторизовались"))
	})
	r.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://shikimori.one/oauth/authorize?client_id="+ClientID+"&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=", 301)
	})
	r.HandleFunc("/anime", func(w http.ResponseWriter, r *http.Request) {

		client := &http.Client{}

		req, _ := http.NewRequest("GET", "https://shikimori.one/api/animes", nil)
		req.Header.Add("Accept", "application/json")
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Authorization", "Bearer "+Token.AccessToken)
		resp, err := client.Do(req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(body)

	})
	r.HandleFunc("/genres", func(w http.ResponseWriter, r *http.Request) {

		client := &http.Client{}

		req, _ := http.NewRequest("GET", "https://shikimori.one/api/genres", nil)
		req.Header.Add("Accept", "application/json")
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Authorization", "Bearer "+Token.AccessToken)
		resp, err := client.Do(req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(body)
	})
	r.HandleFunc("/studios", func(w http.ResponseWriter, r *http.Request) {

		client := &http.Client{}

		req, _ := http.NewRequest("GET", "https://shikimori.one/api/studios", nil)
		req.Header.Add("Accept", "application/json")
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Authorization", "Bearer "+Token.AccessToken)
		resp, err := client.Do(req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(body)

	})
	r.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		pathFrom := query.Get("from")
		if pathFrom == "" {
			pathFrom = "res/cats2.graphml"
		}

		pathTo := query.Get("to")
		if pathTo == "" {
			pathTo = "res/cats_3.graphml"
		}

		var graphml = new(graphml.Graphml)
		err := graphml.Load(pathFrom)
		if err != nil {
			log.Fatal(err)
		}

		var t = tree.NewTree()
		t.FromGraphml(*graphml, &tree.TreeSettings{
			LeavesKnown: false,
		})

		done := make(chan error)
		ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
		go change(ctx, pathFrom, pathTo, t, done)
		err = <-done
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("Finished")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/compare", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		// first := query.Get("first")
		// second := query.Get("second")
		// if first == "" || second == "" {
		// 	w.Write([]byte("Непонятно кого с кем сравнивать"))
		// }

		path := query.Get("path")
		var animes anime.Animes
		animes.Load(path)

		w.Write([]byte("Ищем тайтлы, похожие на" + animes[3].Russian))

		var err = animes[3].Update()
		if err != nil {
			fmt.Println("err is", err)
		}

		var comparator = compare.NewAnimeComparator(animes, nil)
		var list = comparator.EuclideanAll(animes[3])

		for _, v := range list {
			w.Write([]byte("\n\n" + fmt.Sprintf("%v", v.Anime.Russian)))
			w.Write([]byte("\nЕвклидово расстояние:" + fmt.Sprintf("%.6f", v.D)))
			w.Write([]byte("\nОбщая информация:" + fmt.Sprintf("%v", v.Anime)))
		}

	})

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	// FileServer(r)
	server := &http.Server{
		Addr:         Addr,
		Handler:      r,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	log.Println("Server is on localhost" + Addr + "/")
	server.ListenAndServe()
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

/////
