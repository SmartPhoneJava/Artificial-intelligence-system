package main

import (
	"context"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gonum.org/v1/gonum/floats"

	"shiki/internal/anime"
	"shiki/internal/anime/compare"
	"shiki/internal/anime/eatree"
	"shiki/internal/anime/recommend"
	"shiki/internal/anime/shikimori"
	"shiki/internal/anime/tree"
	"shiki/internal/graphml"
	"shiki/internal/models"
	"shiki/internal/page"
	"shiki/internal/score"
	"shiki/internal/score/fs"
	"shiki/internal/utils"
)

const (
	UserAgent    = "Shikimori chrome extension"
	ClientID     = "5427ddff1021ee49d222b276e3f27fc0743896579dc8a79af2c76cfa94781e34"
	ClientSecret = "3ac12dfd95d24dba4ed581f9e47f0ea84bc4d9ed64a9c486be9de0cd0b55b726"

	Addr         = ":2997"
	ReadTimeout  = time.Hour * 3600
	WriteTimeout = time.Hour * 3600
	IdleTimeout  = time.Hour * 3600
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

	var scores = fs.NewScores(api, models.DefaultScoreSettings())
	err = scores.Load("internal/models/users_scores.json")
	if err != nil {
		log.Fatal(err)
	}

	genres, err := models.NewGenres("internal/models/genres.json")
	if err != nil {
		log.Fatal(err)
	}

	var Settings = page.NewPageSettings(genres)

	var myScores = models.NewUserScoreMap(nil)

	router(api, animes, scores, myScores, Settings)
}

func router(
	API shikimori.Api,
	ANIMES anime.AnimesUseCase,
	SCORES score.UseCase,
	myScores models.UserScoreMap,
	Settings *page.PageSettings,
) {
	r := mux.NewRouter()

	file, err := os.Open("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var dists = compare.AnimeAllDistances{Ec: true}

	tpl := template.Must(template.New("item.html").
		Funcs(template.FuncMap{"mul": Mul}).
		Parse(string(b)))

	r.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			var (
				local = ANIMES
			)
			switch Settings.Tabs.CurrentTab {
			case page.Rec:
				{
					var recomendI recommend.RecomendI
					if Settings.Recommend.Kind == "collaborate" {
						recomendI = recommend.NewCollaborativeFiltering(
							ANIMES.Animes(),
							SCORES.Get(),
							myScores,
							*Settings.Recommend,
						)
					} else {
						recomendI = recommend.NewContentOriented(
							ANIMES.Animes(),
							myScores,
							&Settings.SearchSettings.Weights,
						)
					}
					animes, err := recomendI.Recommend()
					if err != nil {
						local = ANIMES
						log.Println("cant get recomendations because", err)
					} else {
						local = anime.NewAnimes(API, animes)
					}
					break
				}
			case page.Compare:
				var (
					animeModels   = ANIMES.Animes()
					comparator    = compare.NewAnimeComparator(animeModels, &Settings.SearchSettings.Weights)
					comparedID    = Settings.CompareWith
					comparedAnime = models.Anime{}
				)

				if comparedID != 0 {
					comparedAnime, _ = ANIMES.FindAnimeByID(int32(comparedID))
				} else {
					comparedAnime = animeModels[0]
				}
				dists = compare.NewAnimeAllDistances(
					comparedAnime,
					comparator.DistanceAll(comparedAnime).Animes(),
				)

				dists.SetFilter(Settings.CompareType)
				log.Println("animes len", len(dists.Animes()))
				local = anime.NewAnimes(API, dists.Animes())
				break
			case page.Fav:
				local = anime.NewAnimes(API, ANIMES.UserAnimes(myScores))
				break
			}

			local.MarkMine(myScores)
			animesFound := local.FindAnimes(Settings.Search)

			var animesFiltered, errText = animesFound.Filter(Settings.SearchSettings)

			tpl.Execute(w, struct {
				Animes    models.Animes
				Page      page.PageSettings
				Distances compare.AnimeAllDistances
				ErrText   string
			}{
				Animes:    animesFiltered.Top(30),
				Page:      *Settings,
				Distances: dists,
				ErrText:   errText,
			})
		})
	r.HandleFunc("/tab_catalog",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabCatalog)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_recomend",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabRecomendations)
			Settings.Recommend = &page.DefaultRecommendSettings
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_favourite",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabFavourite)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_compare",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabCompare)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/set",
		func(w http.ResponseWriter, r *http.Request) {
			if Settings.Tabs.IsCompare {
				compareType := r.URL.Query().Get("compare")

				Settings.CompareType = compareType
			}
			for k, v := range r.URL.Query() {
				switch k {
				case "search":
					Settings.Search = v[0]
					break
				case "rectype":
					Settings.Recommend.Kind = v[0]
					break
				case "users":
					i, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.Recommend.Users = i
					}
					break
				case "percent":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.Recommend.Percent = f
					}
					break
				case "extended":
					Settings.ExtraSearch = !Settings.ExtraSearch
					break
				case "profi":
					Settings.Recommend.WithUserWeights = !Settings.Recommend.WithUserWeights
					break
				case "genres":
					Settings.SearchSettings.SwapGenre(v[0])
					break
				case "kind":
					Settings.SearchSettings.SwapKind(v[0])
					break
				case "oldrating":
					Settings.SearchSettings.SwapOldRating(v[0])
					break
				case "min-year":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MinYear = f
					}
					break
				case "max-year":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MaxYear = f
					}
					break
				case "min-episodes":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MinEpisodes = f
					}
					break
				case "max-episodes":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MaxEpisodes = f
					}
					break
				case "min-duration":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MinDuration = f
					}
					break
				case "max-duration":
					f, err := strconv.Atoi(v[0])
					if err == nil {
						Settings.SearchSettings.MaxDuration = f
					}
					break
				case "min-rating":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.MinRating = f
					}
					break
				case "max-rating":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.MaxRating = f
					}
					break
				case "wkind":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Kind = f
					}
					break
				case "wscore":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Score = f
					}
					break
				case "wepisodes":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Episodes = f
					}
					break
				case "wduration":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Duration = f
					}
					break
				case "wrating":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Rating = f
					}
					break
				case "wyear":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Year = f
					}
					break
				case "wongoing":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Ongoing = f
					}
					break
				case "wstudio":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Studio = f
					}
					break
				case "wgenre":
					f, err := strconv.ParseFloat(v[0], 64)
					if err == nil {
						Settings.SearchSettings.Weights.Genre = f
					}
					break
				}

			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_add",
		func(w http.ResponseWriter, r *http.Request) {
			id, err := utils.RequestInt(r, "id")
			if err != nil {
				log.Println("/favourite post: wrong id given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			score, err := utils.RequestInt(r, "score")
			if err != nil {
				log.Println("/favourite post: wrong score given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			myScores.Add(id, score)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_remove",
		func(w http.ResponseWriter, r *http.Request) {
			id, err := utils.RequestInt(r, "id")
			if err != nil {
				log.Println("/favourite_add: wrong id given")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			myScores.Remove(id)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_remove_all",
		func(w http.ResponseWriter, r *http.Request) {
			myScores.RemoveAll()
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

			graph, err := eatree.NewChartsGraph("assets/res/cats_40.graphml")
			if err != nil {
				log.Fatal("/graph_visual", err)
			}

			if err = graph.Render(w); err != nil {
				log.Println(err)
			}
		})

	r.HandleFunc("/compare",
		func(w http.ResponseWriter, r *http.Request) {
			idInt, err := utils.RequestInt(r, "id")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			Settings.SetTabs("Сравнение")

			Settings.CompareWith = idInt

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/users_update",
		func(w http.ResponseWriter, r *http.Request) {

			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), ReadTimeout)
			go SCORES.Fetch(ctx, 740000, done)
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

func Mul(param1, param2 float64) float64 {
	return floats.Round(param1*param2, 3)
}

///// 473 -> 365
