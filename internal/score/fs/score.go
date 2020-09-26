package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"shiki/internal/models"
)

type Scores struct {
	models.UsersScoreMap
}

func (usm *Scores) Load(path string) error {
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

func (usm Scores) Save(path string) error {
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
