package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"

	"shiki/internal/anime"
	"shiki/internal/anime/compare"
	"shiki/internal/anime/tree"
	"shiki/internal/graphml"
	"shiki/internal/models"
	"shiki/internal/page"
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
	ReadTimeout  = time.Second * 3600
	WriteTimeout = time.Second * 3600
	IdleTimeout  = time.Second * 3600
)

var Token = TokenResponse{
	AccessToken:  "l5xR_Nl-tt4FdT_WX5sxLXnSCQ21B7JPQbI_QjRAYrw",
	Tokentype:    "Bearer",
	ExpiresIn:    86400,
	RefreshToken: "RyBPMUV5g-JwYkEJfL16l1ppxYeRI8CHKnZNShTYP68",
	Scope:        "user_rates comments topics",
	CreatedAt:    1600339489,
}

func main() {
	var animes = anime.NewAnimes(nil)
	animes.Load("assets/res/cats_40.graphml")

	var dists compare.AnimeAllDistances

	router(animes, dists)
}

func router(
	ANIMES anime.AnimesUseCase,
	DISTS compare.AnimeAllDistances,
) {
	r := mux.NewRouter()

	var tpl, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			var local anime.AnimesUseCase
			if page.Settings.Tabs.IsCompare {
				local = anime.NewAnimes(DISTS.Animes())
			} else {
				local = ANIMES
			}
			animesFound := local.FindAnimes(page.Settings.Search)
			log.Printf("Execute start")
			tpl.Execute(w, struct {
				Animes    models.Animes
				Page      page.PageSettings
				Distances compare.AnimeAllDistances
			}{
				Animes:    animesFound.Top(10),
				Page:      page.Settings,
				Distances: DISTS,
			})
		})
	r.HandleFunc("/tab_catalog",
		func(w http.ResponseWriter, r *http.Request) {
			(&page.Settings).SetTabs("Каталог")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_compare",
		func(w http.ResponseWriter, r *http.Request) {
			(&page.Settings).SetTabs("Сравнение")
			var (
				animeModels = ANIMES.Animes()
			)
			var comparator = compare.NewAnimeComparator(animeModels, nil)
			DISTS = comparator.DistanceAll(animeModels[0], true)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/set",
		func(w http.ResponseWriter, r *http.Request) {
			if page.Settings.Tabs.IsCompare {
				compareType := r.URL.Query().Get("compare")
				DISTS.SetFilter(compareType)
			}
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
	r.HandleFunc("/update",
		func(w http.ResponseWriter, r *http.Request) {
			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
			go ANIMES.FetchDetails(ctx, done)
			err := <-done
			if err != nil {
				log.Println("err is", err)
			} else {
				err = ANIMES.Save("assets/res/cats_40.graphml", "assets/res/cats_40.graphml")
				if err != nil {
					log.Println("err is", err)
				}
				log.Println("FINISH")
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})

	r.HandleFunc("/token/{token}",
		func(w http.ResponseWriter, r *http.Request) {

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

			req, err := http.NewRequest(
				"POST",
				"https://shikimori.one/oauth/token",
				bytes.NewBuffer(jsonStr),
			)
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
	r.HandleFunc("/auth",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://shikimori.one/oauth/authorize?client_id="+ClientID+"&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=", 301)
		})
	r.HandleFunc("/anime",
		func(w http.ResponseWriter, r *http.Request) {

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
	r.HandleFunc("/genres",
		func(w http.ResponseWriter, r *http.Request) {

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
	r.HandleFunc("/graph",
		func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			var (
				pathFrom = query.Get("from")
				pathTo   = query.Get("to")
				limit    = query.Get("limit")
			)

			if pathFrom == "" {
				pathFrom = "assets/res/cats2.graphml"
			}

			if pathTo == "" {
				pathTo = "assets/res/cats_40.graphml"
			}

			var limitI, err = strconv.Atoi(limit)
			if limit == "" || err != nil {
				limitI = 40
			}
			var graphml = new(graphml.Graphml)
			err = graphml.Load(pathFrom)
			if err != nil {
				log.Fatal(err)
			}

			var t = tree.NewTree()
			t.FromGraphml(*graphml, &tree.TreeSettings{
				LeavesKnown: false,
			})

			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
			go ANIMES.FetchData(ctx, pathFrom, pathTo, limitI, t, done)
			err = <-done
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}
			log.Printf("Finished")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})

	r.HandleFunc("/compare",
		func(w http.ResponseWriter, r *http.Request) {
			id := r.URL.Query().Get("id")
			(&page.Settings).SetTabs("Сравнение")
			var comparator = compare.NewAnimeComparator(ANIMES.Animes(), nil)
			idInt, err := strconv.Atoi(id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			DISTS = comparator.DistanceAll(ANIMES.FindAnimeByID(int32(idInt)))
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})

	r.PathPrefix("/assets/").Handler(
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir("./assets")),
		),
	)

	server := &http.Server{
		Addr:           Addr,
		Handler:        r,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		IdleTimeout:    IdleTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	log.Println("Server is on localhost" + Addr + "/")
	server.ListenAndServe()
}

/////
