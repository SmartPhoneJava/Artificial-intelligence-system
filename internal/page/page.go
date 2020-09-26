package page

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
		IsCatalog:  tab == "Каталог",
		IsFav:      tab == "Избранное",
		IsRec:      tab == "Рекомендации",
		IsCompare:  tab == "Сравнение",
	}
}

var Settings PageSettings = PageSettings{
	Tag:    "",
	Search: "",
	Tabs:   NewTabs("Каталог"),
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
