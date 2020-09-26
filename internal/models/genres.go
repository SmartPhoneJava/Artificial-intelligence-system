package models

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Genres []Genre

func (genres Genres) Names() []string {
	var arr = make([]string, len(genres))
	for i, genre := range genres {
		arr[i] = genre.Name
	}
	return arr
}

type Genre struct {
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Russian string `json:"russian"`
	Kind    string `json:"kind"`
}

func (g Genres) ToID(russian string) int32 {
	for _, genre := range g {
		if genre.Russian == russian && genre.Kind == "anime" {
			return genre.ID
		}
	}
	return -1
}

func (g Genres) englishType(russian string) string {
	for _, genre := range g {
		if genre.Russian == russian {
			return genre.Name
		}
	}
	return ""
}

func NewGenres(path string) (Genres, error) {
	var genres Genres
	f, err := os.Open(path)
	if err != nil {
		return genres, err
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return genres, err
	}

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &genres)
	if err != nil {
		return genres, err
	}
	return genres, nil
}
