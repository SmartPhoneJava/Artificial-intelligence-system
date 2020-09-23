package main

import (
	"bytes"
	"encoding/json"
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
	"strconv"
	"time"

	"github.com/beevik/etree"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
)

var Token = TokenResponse{
	AccessToken:  "l5xR_Nl-tt4FdT_WX5sxLXnSCQ21B7JPQbI_QjRAYrw",
	Tokentype:    "Bearer",
	ExpiresIn:    86400,
	RefreshToken: "RyBPMUV5g-JwYkEJfL16l1ppxYeRI8CHKnZNShTYP68",
	Scope:        "user_rates comments topics",
	CreatedAt:    1600339489,
}

func findAnime(category, value string, g genres.Genres, s studios.Studios) (anime.Animes, error) {
	client := &http.Client{}

	requests++
	if requests%90 == 0 {
		time.Sleep(time.Minute)
	}

	var query = "https://shikimori.one/api/animes?page=1&limit=20&censored=false&"
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

	fmt.Println("query is", query)
	resp, err := client.Do(req)
	if err != nil {
		return animes, err
	}

	log.Println("resp.Status ", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return animes, err
	}

	//log.Println("aaaaaa", string(body))

	err = json.Unmarshal(body, &animes)
	if err != nil {
		return animes, err
	}
	for _, v := range animes {
		requests++
		if requests%90 == 0 {
			time.Sleep(time.Minute)
		}
		time.Sleep(time.Millisecond * 200)
		v.Update()
	}
	return animes, nil
}

func getAnimeInfo(category, value string, g genres.Genres, s studios.Studios) (anime.Animes, error) {
	client := &http.Client{}

	var query = "https://shikimori.one/api/animes?page=1&limit=20&censored=false&"
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

	fmt.Println("query is", query)
	resp, err := client.Do(req)
	if err != nil {
		return animes, err
	}

	log.Println("resp.Status ", resp.Status)

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

	log.Println("\n nodes", len(nodes))

	var edges = graph.SelectElements("edge")
	var newEdge = edges[0].Copy()
	newEdge.Attr[0].Value = "e" + strconv.FormatInt(int64(len(edges)+1), 10)
	newEdge.Attr[1].Value = sourceID
	newEdge.Attr[2].Value = newID
	graph.AddChild(newEdge)

	log.Println("\n edges", len(edges))
	return nil
}

func change(fromPath, toPath string, tree tree.Tree) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fromPath); err != nil {
		return err
	}

	genres, err := genres.NewGenres("internal/data/genres/genres.json")
	if err != nil {
		return err
	}

	studios, err := studios.NewStudios("internal/data/studios/studios.json")
	if err != nil {
		return err
	}

	fmt.Println("begin2", tree.Categories)
	for _, v := range doc.ChildElements() {
		for _, v1 := range v.ChildElements() {
			if v1.Tag == "graph" {
				for _, v2 := range v1.ChildElements() {
					if v2.Tag == "node" {
						name := tree.NodesNames[v2.Attr[0].Value]
						var category, ok = tree.Categories[name]

						if ok {
							val := v2.Attr[0].Value
							time.Sleep(time.Millisecond * 200)
							animes, err := findAnime(category, name, genres, studios)
							if err != nil {
								return err
							}
							for _, anime := range animes {
								createNode(v1, val, anime)
							}

						}

					}
				}
			}
		}
	}

	f, err := os.OpenFile(toPath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	_, err = doc.WriteTo(f)

	if err != nil {
		return err
	}
	return f.Close()
}

var requests = 0

func main() {
	// animes := new(anime.Animes)
	// animes.Load("res/cats4.graphml")
	// fmt.Println(animes)
	router()
	// fs := http.FileServer(http.Dir("./assets"))
	// http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// log.Println("Listening on :2999...")
	// err := http.ListenAndServe(":2999", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

var ANIMES anime.Animes

func router() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	var tpl, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	// fs := http.FileServer(http.Dir("./assets"))
	// r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// fs := http.FileServer(http.Dir("./assets"))
	// r.Handle("/assets", http.StripPrefix("/assets", fs))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		animeName := chi.URLParam(r, "q")
		ANIMES.Load("")
		animesFound := ANIMES.Search(animeName)
		log.Println("animes are", animesFound)
		tpl.Execute(w, animesFound)
	})
	r.Get("/token/{token}", func(w http.ResponseWriter, r *http.Request) {

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
	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://shikimori.one/oauth/authorize?client_id="+ClientID+"&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=", 301)
	})
	r.Get("/anime", func(w http.ResponseWriter, r *http.Request) {

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
	r.Get("/genres", func(w http.ResponseWriter, r *http.Request) {

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
	r.Get("/studios", func(w http.ResponseWriter, r *http.Request) {

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
	r.Get("/graph", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		pathFrom := query.Get("from")
		if pathFrom == "" {
			pathFrom = "res/cats2.graphml"
		}

		pathTo := query.Get("to")
		if pathTo == "" {
			pathTo = "res/cats5.graphml"
		}

		var graphml = new(graphml.Graphml)
		err := graphml.Load(pathFrom)
		if err != nil {
			log.Fatal(err)
		}

		var t = tree.NewTree()
		t.FromGraphml(*graphml, &tree.TreeSettings{false})
		fmt.Println("begin")
		err = change(pathFrom, pathTo, t)
		if err != nil {
			log.Fatal(err)
		}
	})
	r.Get("/compare", func(w http.ResponseWriter, r *http.Request) {
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

	r.Get("/find/{query}", func(w http.ResponseWriter, r *http.Request) {

		client := &http.Client{}

		query := chi.URLParam(r, "query")

		req, _ := http.NewRequest("GET", "https://shikimori.one/api/animes?page=1&limit=45&"+query, nil)
		// req.Header.Add("Accept", "application/json")
		// req.Header.Set("User-Agent", UserAgent)
		// req.Header.Set("Authorization", "Bearer "+Token.AccessToken)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("resp.Status ", resp.Status)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(body)
		if err != nil {
			log.Fatal(err)
		}

	})

	FileServer(r)
	server := &http.Server{
		Addr:         ":2999",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	server.ListenAndServe()
}

// FileServer is serving static files.
func FileServer(router *chi.Mux) {
	root := "./assets"
	fs := http.FileServer(http.Dir(root))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
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
