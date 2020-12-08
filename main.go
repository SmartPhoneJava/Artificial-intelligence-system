package main

import (
	"log"
	"net/http"
	"time"

	"shiki/internal/anime"
	"shiki/internal/anime/shikimori"
	"shiki/internal/dialog"
	"shiki/internal/models"
	"shiki/internal/page"
	"shiki/internal/router"
	"shiki/internal/score/fs"

	_ "github.com/jinzhu/gorm/dialects/postgres"
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

	var (
		Settings = page.NewPageSettings(genres)

		myScores   = models.NewUserScoreMap(nil)
		toDialog   = make(chan string, 1)
		fromDialog = make(chan dialog.NLPResponse, 1)
		messages   = make(dialog.Messages, 0)
	)
	(&messages).Add("Привет, поболтаем? Ты можешь общаться со мной через текстовое поле над моей головой!", true)

	r := router.New(api, animes, scores, myScores,
		Settings, ReadTimeout, messages, UserAgent,
		ClientID, toDialog, fromDialog)

	server := &http.Server{
		Addr:           Addr,
		Handler:        r,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		IdleTimeout:    IdleTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	log.Println("Server is on localhost" + Addr + "/")
	go dialog.Run(toDialog, fromDialog)
	server.ListenAndServe()
}

///// 473 -> 365
