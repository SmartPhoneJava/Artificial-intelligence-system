package page

const COLLABORATE = "collaborate"
const CONTENT = "content"

const TabRecomendations = "Рекомендации"
const TabCatalog = "Каталог"
const TabFavourite = "Избранное"
const TabCompare = "Сравнение"

var DefaultRecommendSettings = RecommendSettings{
	Kind:    COLLABORATE,
	Users:   10,
	Percent: 0.5,
}
