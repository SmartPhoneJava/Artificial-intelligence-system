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
	"shiki/internal/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const UsersCount = 740000

func New(
	API shikimori.Api,
	Input recommend.Input,
	Settings *page.PageSettings,
	readTimeout time.Duration,
	userAgent, clientID string,
	comm dialog.Communication,
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
	var dists = compare.AnimeAllDistances{Ec: true}

	tpl := template.Must(template.New("item.html").
		Funcs(template.FuncMap{
			"mul": Mul,
			"ec":  Ec(dists),
			"mc":  Mc(dists),
			"kc":  Kc(dists),
			"dc":  Dc(dists),
			"cc":  Cc(dists),
			"tc":  Tc(dists),
			"safeHTML": func(s interface{}) template.HTML {
				return template.HTML(fmt.Sprint(s))
			},
			"ShowBranch": atemplate.ShowBranch,
		}).ParseFiles(
		"assets/templates/branch.html",
		"assets/templates/marks.html",
		"assets/templates/anime_info.html",
		"assets/templates/extra-search.html",
		"assets/templates/collaborate_users.html",
		"assets/templates/compare_type.html",
		"assets/templates/compare_info.html",
		"assets/templates/score_mine.html",
		"assets/templates/score_their.html",
		"index.html",
	))

	r.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			var (
				local = Input.Animes
			)
			switch Settings.Tabs.CurrentTab {
			case page.Rec:
				{
					var recomendI recommend.RecomendI
					log.Println("/recommend", Settings.Recommend.Kind)
					if Settings.Recommend.Kind == "collaborate" {
						recomendI = recommend.NewCollaborativeFiltering(
							Input,
							*Settings.Recommend,
						)
					} else {
						recomendI = recommend.NewContentOriented(
							Input,
							&Settings.SearchSettings.Weights,
						)
					}
					animes, err := recomendI.Recommend()
					if err != nil {
						local = Input.Animes
						log.Println("cant get recomendations because", err)
					} else {
						local = anime.NewAnimes(API, animes)
					}
					break
				}
			case page.Compare:
				var (
					animeModels   = Input.Animes.Animes()
					comparedAnime = models.Anime{}
					comparedID    = int32(Settings.CompareWith)
				)
				if comparedID != 0 {
					comparedAnime, _ = Input.Animes.FindAnimeByID(comparedID)
				} else {
					comparedAnime = animeModels[0]
				}

				var comparator = compare.NewAnimeComparator(
					animeModels.AllExcept(comparedAnime),
					&Settings.SearchSettings.Weights,
				)

				dists = compare.NewAnimeAllDistances(
					comparedAnime,
					comparator.DistanceAll(comparedAnime).Animes(),
				)

				dists.SetFilter(Settings.CompareType)
				local = anime.NewAnimes(API, dists.Animes())
				break

			case page.Fav:
				myAnimes := Input.Animes.UserAnimes(Input.MyScores)
				local = anime.NewAnimesNoApi(myAnimes)
				break
			}

			local.MarkMine(Input.MyScores)
			animesFound := local.FilterByName(Settings.Search)

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
				Messages:  *comm.Messages,
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
	r.HandleFunc("/set", routeSet(Input, Settings, comm))
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

			Input.MyScores.Add(id, score)
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

			Input.MyScores.Remove(id)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/favourite_remove_all",
		func(w http.ResponseWriter, r *http.Request) {
			Input.MyScores.RemoveAll()
			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	r.HandleFunc("/update",
		func(w http.ResponseWriter, r *http.Request) {
			done := make(chan error)
			ctx, _ := context.WithTimeout(r.Context(), readTimeout)
			go Input.Animes.FetchDetails(ctx, userAgent, done)
			err := <-done
			if err != nil {
				log.Println("err is", err)
			} else {
				err = Input.Animes.Save(
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
			go Input.Animes.FetchData(ctx, pathFrom, pathTo, limitI, t, done)
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
			go Input.AllScores.Fetch(ctx, UsersCount, done)
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

func routeSet(
	input recommend.Input,
	settings *page.PageSettings,
	comm dialog.Communication,
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
					talkWithUser(
						v[0],
						comm,
						input,
						settings,
					)
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

func talkWithUser(
	message string,
	comm dialog.Communication,
	input recommend.Input,
	settings *page.PageSettings,
) {
	comm.Messages.Add(message, false)
	if message != "" {
		comm.ToDialog <- message
		var response = <-comm.FromDialog
		msg, isObjectIntent := dialog.HandleResponseText(
			response,
			input.Animes,
			input.MyScores,
		)

		if isObjectIntent {
			obj, err := dialog.HandleResponseObjects(
				response,
				input,
				settings.SearchSettings.Genres,
			)
			if err != nil {
				log.Println("Err in isObjectIntent", err.Error())
				comm.Messages.Add("Извини, кажется я не знаю, что тебе ответить, перефразируй своё сообщение пожалуйста.", true)
			} else {
				comm.Messages.AddWithAnime(obj.Message, obj.Animes)
			}
		} else {
			comm.Messages.Add(msg, true)
		}
	}
}

// 541
