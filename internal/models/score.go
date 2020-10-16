package models

var UsersScores UsersScoreMap

type ScoreSettings struct {
	SavePath string `json:"save_path"`
}

func DefaultScoreSettings() *ScoreSettings {
	return &ScoreSettings{
		SavePath: "internal/models/users_scores.json",
	}
}

type UsersScoreMap []UserScoreMap

func (d UsersScoreMap) Len() int           { return len(d) }
func (d UsersScoreMap) Less(i, j int) bool { return d[i].D < d[j].D }
func (d UsersScoreMap) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

type UserScoreMap struct {
	User   int         `json:"user"`
	Scores map[int]int `json:"scores"`
	D      float64
}

func (usm UserScoreMap) Add(id, score int) {
	usm.Scores[id] = score
}

func (usm UserScoreMap) Remove(id int) {
	delete(usm.Scores, id)
}

func (usm *UserScoreMap) RemoveAll() {
	usm.Scores = make(map[int]int, 0)
}

func (usm UserScoreMap) ExceptMine(allAnimes Animes) Animes {
	var newAnimes = make([]Anime, 0)
	for _, a := range allAnimes {
		if usm.Scores[int(a.ID)] == 0 {
			newAnimes = append(newAnimes, a)
		}
	}
	return newAnimes

}

func NewUserScoreMap(scores Scores) UserScoreMap {
	if len(scores) == 0 {
		return UserScoreMap{
			Scores: make(map[int]int, 0),
		}
	}
	var m = make(map[int]int, len(scores))
	for _, score := range scores {
		if score.Score > 0 {
			m[score.TargetID] = score.Score
		}
	}
	return UserScoreMap{
		User:   scores[0].UserID,
		Scores: m,
	}
}

type Scores []Score

type Score struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	TargetID   int    `json:"target_id"`
	TargetType string `json:"target_type"`
	Status     string `json:"status"`
	Score      int    `json:"score"`
}
