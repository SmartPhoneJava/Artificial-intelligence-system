package router

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"shiki/internal/anime"
	"shiki/internal/anime/compare"
	"shiki/internal/anime/eatree"
	"shiki/internal/anime/recommend"
	"shiki/internal/anime/shikimori"
	"shiki/internal/anime/tree"
	"shiki/internal/atemplate"
	"shiki/internal/dialog"
	"shiki/internal/graphml"
	"shiki/internal/models"
	"shiki/internal/page"
	"shiki/internal/score"
	"shiki/internal/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gonum.org/v1/gonum/floats"
)

const UsersCount = 740000

func New(
	API shikimori.Api,
	ANIMES anime.AnimesUseCase,
	SCORES score.UseCase,
	myScores models.UserScoreMap,
	Settings *page.PageSettings,
	readTimeout time.Duration,
	Messages dialog.Messages,
	userAgent, clientID string,
	toDialog chan string,
	fromDialog chan dialog.NLPResponse,
) *mux.Router {
	r := mux.NewRouter()

	log.Println("readTimeout", readTimeout)

	file, err := os.Open("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// b, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var dists = compare.AnimeAllDistances{Ec: true}

	// tmpl := make(map[string]*template.Template)
	// tmpl["index.html"] = template.Must(template.ParseFiles("branch.tmpl", "index.html"))
	//tpl := template.Must(template.ParseFiles("index.html", "home.html"))
	tpl := template.Must(template.New("item.html").
		Funcs(template.FuncMap{
			"mul": Mul,
			"safeHTML": func(s interface{}) template.HTML {
				return template.HTML(fmt.Sprint(s))
			},
			"ShowBranch": atemplate.ShowBranch,
		}).ParseFiles(
		"assets/templates/branch.html",
		"assets/templates/marks.html",
		"assets/templates/anime_info.html",
		"assets/templates/extra-search.html",
		"home.html",
		"base.html",
		"index.html",
	))

	/*
				//tmpl["branch.html"] = template.Must(template.ParseFiles("branch.tmpl", "index.html"))
		tpl := template.New("item.html")
		tpl = tpl.Funcs(template.FuncMap{
			"mul": Mul,
			"safeHTML": func(s interface{}) template.HTML {
				return template.HTML(fmt.Sprint(s))
			},
			"ShowBranch": atemplate.ShowBranch,
		})
		tpl = template.Must(tpl.ParseFiles("index.html", "branch.tmpl"))


	*/

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

			tpl.ExecuteTemplate(w, "base", struct {
				Animes    models.Animes
				Page      page.PageSettings
				Distances compare.AnimeAllDistances
				ErrText   string
				Desc      template.HTML
				Messages  dialog.Messages
			}{
				Animes:    animesFiltered.Top(30),
				Page:      *Settings,
				Distances: dists,
				ErrText:   errText,
				Desc:      template.HTML("<p>Paragraph</p>"),
				Messages:  Messages,
			})
		})
	r.HandleFunc("/tab_catalog",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabCatalog)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/tab_smart",
		func(w http.ResponseWriter, r *http.Request) {
			Settings.SetTabs(page.TabSmart)
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
	r.HandleFunc("/set", routeSet(
		ANIMES,
		myScores,
		SCORES,
		Settings,
		&Messages,
		toDialog,
		fromDialog,
	))
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
			ctx, _ := context.WithTimeout(r.Context(), readTimeout)
			go ANIMES.FetchDetails(ctx, userAgent, done)
			err := <-done
			if err != nil {
				log.Println("err is", err)
			} else {
				err = ANIMES.Save(
					"assets/res/cats_40.graphml",
					"assets/res/cats_40.graphml",
				)
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
			http.Redirect(w, r,
				"https://shikimori.one/oauth/authorize?client_id="+
					clientID+
					"&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=",
				301)
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
			ctx, _ := context.WithTimeout(r.Context(), readTimeout)
			go ANIMES.FetchData(ctx, pathFrom, pathTo, limitI, t, done)
			err = <-done
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}
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
			ctx, _ := context.WithTimeout(r.Context(), readTimeout)
			go SCORES.Fetch(ctx, UsersCount, done)
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

	return r
}

func Mul(param1, param2 float64) float64 {
	return floats.Round(param1*param2, 3)
}

func routeSet(
	animes anime.AnimesUseCase,
	myScores models.UserScoreMap,
	allScores score.UseCase,
	settings *page.PageSettings,
	messages *dialog.Messages,
	toDialog chan string,
	fromDialog chan dialog.NLPResponse,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if settings.Tabs.IsCompare {
			compareType := r.URL.Query().Get("compare")

			settings.CompareType = compareType
		}
		for k, v := range r.URL.Query() {
			switch k {
			case "search":
				if settings.Tabs.CurrentTab == page.Smart {
					messages.Add(v[0], false)
					if v[0] != "" {
						toDialog <- v[0]
						var response = <-fromDialog
						msg, isObjectIntent := dialog.HandleResponseText(
							response,
							animes,
							myScores,
						)

						if isObjectIntent {
							obj, err := dialog.HandleResponseObjects(
								response,
								animes,
								myScores,
								allScores,
								settings.SearchSettings.Genres,
							)
							if err != nil {
								log.Println("Err in isObjectIntent", err.Error())
								messages.Add("Извини, кажется я не знаю, что тебе ответить, перефразируй своё сообщение пожалуйста.", true)
							} else {
								messages.AddWithAnime(obj.Message, obj.Animes)
							}
						} else {
							messages.Add(msg, true)
						}
					}
				} else {
					settings.Search = v[0]
				}
			case "rectype":
				settings.Recommend.Kind = v[0]
			case "users":
				i, err := strconv.Atoi(v[0])
				if err == nil {
					settings.Recommend.Users = i
				}
			case "percent":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.Recommend.Percent = f
				}
			case "extended":
				settings.ExtraSearch = !settings.ExtraSearch
			case "profi":
				settings.Recommend.WithUserWeights = !settings.Recommend.WithUserWeights
			case "genres":
				settings.SearchSettings.SwapGenre(v[0])
			case "kind":
				settings.SearchSettings.SwapKind(v[0])
			case "oldrating":
				settings.SearchSettings.SwapOldRating(v[0])
			case "min-year":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MinYear = f
				}
			case "max-year":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MaxYear = f
				}
			case "min-episodes":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MinEpisodes = f
				}
			case "max-episodes":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MaxEpisodes = f
				}
			case "smart-mode":
				log.Println("smart-mode", v[0] == "true")
				settings.SearchSettings.SmartMode = v[0] == "true"
			case "min-duration":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MinDuration = f
				}
			case "max-duration":
				f, err := strconv.Atoi(v[0])
				if err == nil {
					settings.SearchSettings.MaxDuration = f
				}
			case "min-rating":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.MinRating = f
				}
			case "max-rating":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.MaxRating = f
				}
			case "wkind":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Kind = f
				}
			case "wscore":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Score = f
				}
			case "wepisodes":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Episodes = f
				}
			case "wduration":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Duration = f
				}
			case "wrating":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Rating = f
				}
			case "wyear":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Year = f
				}
			case "wongoing":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Ongoing = f
				}
			case "wstudio":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Studio = f
				}
			case "wgenre":
				f, err := strconv.ParseFloat(v[0], 64)
				if err == nil {
					settings.SearchSettings.Weights.Genre = f
				}
			}

		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
