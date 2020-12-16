package page

import (
	"shiki/internal/models"
)

const (
	Catalog = "Каталог"
	Fav     = "Избранное"
	Rec     = "Рекомендации"
	Compare = "Сравнение"
	Smart   = "Ассистент"
)

type Tabs struct {
	CurrentTab                                  string
	IsCatalog, IsFav, IsRec, IsCompare, IsSmart bool
}

type Panels struct {
	CurrentPanel string
}

func NewPanels(panel string) Panels {
	return Panels{
		CurrentPanel: panel,
	}
}

func NewTabs(tab string) Tabs {
	return Tabs{
		CurrentTab: tab,
		IsCatalog:  tab == Catalog,
		IsFav:      tab == Fav,
		IsRec:      tab == Rec,
		IsCompare:  tab == Compare,
		IsSmart:    tab == Smart,
	}
}

func NewPageSettings(genres models.Genres) *PageSettings {
	return &PageSettings{
		Search:         "",
		CompareType:    "e",
		CompareWith:    0,
		Tabs:           NewTabs(Catalog),
		Panels:         NewPanels("Все"),
		Recommend:      &RecommendSettings{},
		SearchSettings: NewSearchSettings(genres),
	}
}

func (pg *PageSettings) SetTabs(tab string) {
	pg.Tabs = NewTabs(tab)
}

type RecommendSettings struct {
	Kind            string
	Users           int
	Percent         float64
	WithUserWeights bool
}

func toCompressedArr(genres models.Genres) []models.GenresMarked {
	var (
		arrOfArrays = []models.GenresMarked{}
		array6      = []models.GenreMarked{}
		counter     = 0
		textLimit   = 140
	)
	for i := 0; i < len(genres); i++ {
		if genres[i].Kind == "manga" {
			continue
		}
		counter += len(genres[i].Russian)
		if counter > textLimit {
			arrOfArrays = append(arrOfArrays, array6)
			array6 = []models.GenreMarked{{
				Genre:  genres[i],
				Marked: true,
			}}
			counter = 0
		} else {
			array6 = append(array6, models.GenreMarked{
				Genre:  genres[i],
				Marked: true,
			})
		}
	}
	return arrOfArrays
}

func NewSearchSettings(genres models.Genres) *models.SearchSettings {
	searchSettings := models.NewSimpleSearchSettings()
	searchSettings.Genres = toCompressedArr(genres)
	return searchSettings
}

type PageSettings struct {
	Search         string
	ExtraSearch    bool
	CompareType    string
	CompareWith    int
	SearchSettings *models.SearchSettings
	Recommend      *RecommendSettings
	Tabs           Tabs
	Panels         Panels
}
