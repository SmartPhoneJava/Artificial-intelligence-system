package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/gorilla/mux"

	"shiki/internal/anime"
	"shiki/internal/anime/compare"
	"shiki/internal/anime/eatree"
	"shiki/internal/anime/shikimori"
	"shiki/internal/anime/tree"
	"shiki/internal/graphml"
	"shiki/internal/models"
	"shiki/internal/page"
	"shiki/internal/score"
	"shiki/internal/score/fs"
)

const (
	UserAgent    = "Shikimori chrome extension"
	ClientID     = "5427ddff1021ee49d222b276e3f27fc0743896579dc8a79af2c76cfa94781e34"
	ClientSecret = "3ac12dfd95d24dba4ed581f9e47f0ea84bc4d9ed64a9c486be9de0cd0b55b726"

	Addr         = ":2997"
	ReadTimeout  = time.Second * 3600
	WriteTimeout = time.Second * 3600
	IdleTimeout  = time.Second * 3600
)

var Token = shikimori.TokenResponse{
	AccessToken:  "l5xR_Nl-tt4FdT_WX5sxLXnSCQ21B7JPQbI_QjRAYrw",
	Tokentype:    "Bearer",
	ExpiresIn:    86400,
	RefreshToken: "RyBPMUV5g-JwYkEJfL16l1ppxYeRI8CHKnZNShTYP68",
	Scope:        "user_rates comments topics",
	CreatedAt:    1600339489,
}

func main() {
	var api = shikimori.NewApi(UserAgent, ClientID, ClientSecret, true, WriteTimeout/2)

	var animes = anime.NewAnimes(api, nil)
	err := animes.Load("assets/res/cats_40.graphml")
	if err != nil {
		log.Fatal(err)
	}

	var dists compare.AnimeAllDistances

	var scores = fs.NewScores(api, models.DefaultScoreSettings())
	err = scores.Load("internal/models/users_scores.json")
	if err != nil {
		log.Fatal(err)
	}

	var myScores = models.NewUserScoreMap(nil)

	router(api, animes, dists, scores, myScores)
}

func router(
	API shikimori.Api,
	ANIMES anime.AnimesUseCase,
	DISTS compare.AnimeAllDistances,
	SCORES score.UseCase,
	myScores models.UserScoreMap,
) {
	r := mux.NewRouter()

	var tpl, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			var (
				local = ANIMES
			)
			switch page.Settings.Tabs.CurrentTab {
			case page.Rec:
				{
					animes, err := compare.NewCollaborativeFiltering(
						ANIMES.Animes(),
						SCORES.Get(),
						SCORES.Get()[0],
					).Recomend(100)
					if err != nil {
						local = ANIMES
						log.Println("cant get recomendations because", err)
					} else {
						local = anime.NewAnimes(API, animes)
					}
					break
				}
			case page.Compare:
				local = anime.NewAnimes(API, DISTS.Animes())
				break
			case page.Fav:
				local = anime.NewAnimes(API, ANIMES.UserAnimes(myScores))
				break
			}

			animesFound := local.FindAnimes(page.Settings.Search)
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
	r.HandleFunc("/tab_recomend",
		func(w http.ResponseWriter, r *http.Request) {
			(&page.Settings).SetTabs("Рекомендации")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_favourite",
		func(w http.ResponseWriter, r *http.Request) {
			(&page.Settings).SetTabs("Избранное")
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
	r.HandleFunc("/favourite",
		func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			id := query.Get("id")
			if id == "" {
				log.Println("/favourite post: no id given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			idI, err := strconv.Atoi(id)
			if err != nil {
				log.Println("/favourite post: wrong id given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			score := query.Get("score")
			if id == "" {
				log.Println("/favourite post: no score given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			scoreI, err := strconv.Atoi(score)
			if err != nil {
				log.Println("/favourite post: wrong score given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			myScores.Scores[idI] = scoreI

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_add",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("url is", r.URL)
			query := r.URL.Query()

			id := query.Get("id")
			idI, err := strconv.Atoi(id)
			if err != nil {
				log.Println("/favourite_add: wrong id given", id)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			score := query.Get("score")
			scoreI, err := strconv.Atoi(score)
			if err != nil {
				log.Println("/favourite_add: wrong score given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			myScores.Add(idI, scoreI)
			ANIMES.MarkMine(myScores)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_remove",
		func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			id := query.Get("id")
			idI, err := strconv.Atoi(id)
			if err != nil {
				log.Println("/favourite_add: wrong id given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			myScores.Remove(idI)
			ANIMES.MarkMine(myScores)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/update",
		func(w http.ResponseWriter, r *http.Request) {
			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
			go ANIMES.FetchDetails(ctx, UserAgent, done)
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
		API.Auth(
			func(r *http.Request) string {
				return mux.Vars(r)["token"]
			},
		))
	r.HandleFunc("/auth",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://shikimori.one/oauth/authorize?client_id="+ClientID+"&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=", 301)
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
	r.HandleFunc("/graph_visual",
		func(w http.ResponseWriter, r *http.Request) {

			graph := charts.NewGraph()
			graph.SetGlobalOptions(
				charts.ColorOpts{"green", "red", "blue"},
				charts.TitleOpts{
					Title: "Вернуться назад",
					Link:  "/",
				},
				charts.LegendOpts{Right: "20%"},
				charts.ToolboxOpts{Show: true},
				charts.InitOpts{
					PageTitle: "Визуализация классификации аниме",
					Width:     "720px", Height: "750px",
					BackgroundColor: "#f5f5dc"},
				// charts.DataZoomOpts{XAxisIndex: []int{0}, Start: 50, End: 100},
			)

			var eadoc, err = eatree.NewEdoc("assets/res/cats_40.graphml")
			if err != nil {
				log.Fatal(err)
			}
			graphNodes := make([]charts.GraphNode, 0)
			graphLinks := make([]charts.GraphLink, 0)

			var chooseColor = func(v int) string {
				switch v {
				case 1:
					return "brown"
				case 2:
					return "orange"
				case 3:
					return "blue"
				case 4:
					return "green"
				case 5:
					return "red"
				}
				return "black"
			}

			var chooseForm = func(v int) string {
				// switch v {
				// case 1:
				// 	return "pin"
				// case 2:
				// 	return "rect"
				// case 3:
				// 	return "roundRect"
				// case 4:
				// 	return "diamond"
				// case 5:
				// 	return "triangle"
				// }
				return "diamond"
			}

			var chooseSize = func(v int) interface{} {
				switch v {
				case 1:
					return []int{50, 50}
				case 2:
					return []int{40, 40}
				case 3:
					return []int{30, 30}
				case 4:
					return []int{20, 20}
				case 5:
					return []int{15, 15}
				}
				return []int{10, 10}
			}

			var m = make(map[string]bool)

			for _, v := range eadoc.Leaves {
				ok := true
				var key = v.NodeID

				for ok {
					if !m[key] {
						var depth = eadoc.Tree.Depth(key)
						graphNodes = append(
							graphNodes,
							charts.GraphNode{
								Name:       key,
								Symbol:     chooseForm(depth),
								SymbolSize: chooseSize(depth),
								ItemStyle: charts.ItemStyleOpts{
									Color: chooseColor(depth),
								},
							})
					}
					m[key] = true
					nods := eadoc.Tree.NodesUp[key]
					if len(nods) == 0 {
						break
					}

					graphLinks = append(graphLinks, charts.GraphLink{Source: nods[0], Target: key})
					key = nods[0]
				}
			}

			graph.Add("graph", graphNodes, graphLinks,
				charts.GraphOpts{Roam: true, FocusNodeAdjacency: true, Force: charts.GraphForce{
					Repulsion: 100,
				}},
				charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"},
					ItemStyle: charts.ItemStyleOpts{Color: "yellow"}},
				charts.LineStyleOpts{Curveness: 0.2})

			err = graph.Render(w)
			if err != nil {
				log.Println(err)
			}
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
	r.HandleFunc("/users_update",
		func(w http.ResponseWriter, r *http.Request) {

			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
			go SCORES.Fetch(ctx, 1000, done)
			err = <-done
			if err != nil {
				log.Println("users_update err is", err)
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}
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
