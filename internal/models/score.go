package models

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var UsersScores UsersScoreMap

type UsersScoreMap []UserScoreMap

func (usm *UsersScoreMap) Load(path string) error {
	if path == "" {
		path = "internal/models/users_scores.json"
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, usm)
	if err != nil {
		return err
	}
	return nil
}

func (usm UsersScoreMap) Save(path string) error {
	if path == "" {
		path = "internal/models/users_scores_2.json"
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	bytesS, err := json.Marshal(usm)
	if err != nil {
		return err
	}
	_, err = f.Write(bytesS)
	if err != nil {
		return err
	}
	return nil
}

type UserScoreMap struct {
	User   User          `json:"user"`
	Scores map[int32]int `json:"scores"`
}

func NewUserScoreMap(scores Scores) UserScoreMap {
	if len(scores) == 0 {
		return UserScoreMap{}
	}
	var m = make(map[int32]int, len(scores))
	for _, score := range scores {
		if score.Score > 0 {
			m[score.Anime.ID] = score.Score
		}
	}
	return UserScoreMap{
		User:   scores[0].User,
		Scores: m,
	}

}

type Scores []Score

type Score struct {
	User  User  `json:"user"`
	Anime Anime `json:"anime"`
	Score int   `json:"score"`
}
