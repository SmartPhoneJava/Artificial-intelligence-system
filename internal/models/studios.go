package models

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Studios []Studio

func (studios Studios) Names() []string {
	var arr = make([]string, len(studios))
	for i, studio := range studios {
		arr[i] = studio.Name
	}
	return arr
}

type Studio struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	FilteredName string `json:"filtered_name"`
	Real         bool   `json:"real"`
}

func (studios Studios) ToID(name string) int32 {
	for _, studio := range studios {
		if studio.Name == name {
			return studio.ID
		}
	}
	return -1
}

func NewStudios(path string) (Studios, error) {
	var studios Studios
	f, err := os.Open(path)
	if err != nil {
		return studios, err
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return studios, err
	}

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &studios)
	if err != nil {
		return studios, err
	}
	return studios, nil
}
