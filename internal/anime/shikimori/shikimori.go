package shikimori

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"shiki/internal/models"
	"shiki/internal/utils"
	"time"
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

type Api struct {
	token                             TokenResponse
	userAgent, clientID, clientSecret string
	ignoreAuth                        bool
	client                            *http.Client
}

func NewApi(
	userAgent, clientID, clientSecret string,
	ignoreAuth bool,
	timeout time.Duration,
) Api {
	return Api{
		userAgent:    userAgent,
		clientID:     clientID,
		clientSecret: clientSecret,
		ignoreAuth:   ignoreAuth,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (api Api) Auth(
	getToken func(r *http.Request) string,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonStr, err := json.Marshal(TokenRequest{
			GrantType:    "authorization_code",
			ClientID:     api.clientID,
			ClientSecret: api.clientSecret,
			Code:         getToken(r),
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
		req.Header.Set("User-Agent", api.userAgent)
		req.Header.Set("Content-Type", "application/json")

		resp, err := api.client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, &api.token)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (api Api) GetScores(
	userID int32,
) (models.Scores, error) {

	req, err := http.NewRequest(
		"GET",
		"https://shikimori.one/api/v2/user_rates",
		nil,
	)
	if err != nil {
		return models.Scores{}, err
	}
	q := req.URL.Query()
	q.Add("user_id", utils.String(userID))
	q.Add("target_type", "Anime")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", api.userAgent)

	scores := models.Scores{}
	resp, err := api.client.Do(req)
	if err != nil {
		return scores, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return scores, utils.Err429
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scores, err
	}

	err = json.Unmarshal(body, &scores)
	if err != nil {
		return scores, err
	}

	return scores, err
}

func (api Api) GetAnimes(
	limit int32,
	genre string,
	studio string,
) (models.Animes, error) {

	req, err := http.NewRequest(
		"GET",
		"https://shikimori.one/api/animes",
		nil,
	)
	if err != nil {
		return models.Animes{}, err
	}
	q := req.URL.Query()
	q.Add("page", "1")
	q.Add("limit", utils.String(limit))
	q.Add("order", "ranked")
	if genre != "" {
		q.Add("genre", genre)
	}
	if studio != "" {
		q.Add("studio", studio)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", api.userAgent)

	var animes models.Animes

	resp, err := api.client.Do(req)
	if err != nil {
		return animes, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return animes, utils.Err429
	} else if resp.StatusCode != http.StatusOK {
		return animes, errors.New("Wrong status:" + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return animes, err
	}

	if err = json.Unmarshal(body, &animes); err != nil {
		return animes, err
	}
	return animes, nil
}
