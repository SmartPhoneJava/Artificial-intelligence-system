package page

const (
	Catalog = "Каталог"
	Fav     = "Избранное"
	Rec     = "Рекомендации"
	Compare = "Сравнение"
)

type Tabs struct {
	CurrentTab                         string
	IsCatalog, IsFav, IsRec, IsCompare bool
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
	}
}

var Settings PageSettings = PageSettings{
	Tag:    "",
	Search: "",
	Tabs:   NewTabs(Catalog),
	Panels: NewPanels("Все"),
}

func (pg *PageSettings) SetTabs(tab string) {
	pg.Tabs = NewTabs(tab)
}

type PageSettings struct {
	Tag, Search string
	Tabs        Tabs
	Panels      Panels
}
