package fs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"shiki/internal/anime/shikimori"
	"shiki/internal/models"
	"shiki/internal/score"
	"shiki/internal/utils"
	"strconv"
)

type Scores struct {
	models.UsersScoreMap
	settings *models.ScoreSettings
	api      shikimori.Api
}

func NewScores(
	api shikimori.Api,
	settings *models.ScoreSettings,
) score.UseCase {
	var scores = Scores{
		api:           api,
		UsersScoreMap: []models.UserScoreMap{},
	}
	scores.SetSettings(settings)
	return &scores
}

func (usm *Scores) SetSettings(settings *models.ScoreSettings) {
	if settings == nil {
		settings = models.DefaultScoreSettings()
	}
	usm.settings = settings
}

func (usm Scores) Get() models.UsersScoreMap {
	return usm.UsersScoreMap
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
		path = usm.settings.SavePath
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bytesS, err := json.Marshal(usm)
	if err != nil {
		return err
	}
	_, err = file.Write(bytesS)
	if err != nil {
		return err
	}
	return nil
}

func (usm *Scores) Fetch(
	ctx context.Context,
	users int32,
	done chan error,
) {
	if users < 1 {
		done <- nil
		return
	}

	var (
		min            int32 = 0
		i              int32
		cancelledUsers int32 = 0
		saveName             = "internal/models/users_scores0.json"
	)

	var newScores = []models.UserScoreMap{}

	for i = min; i < min+users; i++ {

		var userScore models.UserScoreMap
		err := utils.MakeAction(ctx, func() error {
			scores, err := usm.api.GetScores(i)
			if err != nil {
				return err
			}
			if len(scores) != 0 {
				userScore = models.NewUserScoreMap(scores)
				newScores = append(newScores, userScore)
			} else {
				cancelledUsers++
			}
			return nil
		})
		if err != nil {
			done <- err
			return
		}
		log.Printf("User loaded %d/%d",
			i+1-min-cancelledUsers,
			users-cancelledUsers)
		if i%10000 == 0 {
			saveName = "internal/models/users_scores" + strconv.Itoa(int(i)) + ".json"
		}
		if i%10 == 0 {
			usm.UsersScoreMap = newScores
			err = usm.Save(saveName)
			if err != nil {
				done <- err
				return
			}
			log.Printf("Saved")
		}
	}
	usm.UsersScoreMap = newScores
	err := usm.Save("")
	if err != nil {
		done <- err
		return
	}
	log.Printf("All users loaded")
	done <- nil
}
