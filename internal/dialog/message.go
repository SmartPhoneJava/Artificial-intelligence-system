package dialog

import (
	"shiki/internal/models"
	"time"
)

// Message - сообщение
type Message struct {
	FromSystem bool
	Message    string
	Time       string
	NeedAnime  bool
	Animes     models.Animes
}

// Messages - массив сообщений
type Messages []Message

// Add добавить сообщение
func (msgs *Messages) Add(text string, system bool) {
	*msgs = append([]Message{{
		FromSystem: system,
		Message:    text,
		Time:       time.Now().Format("15:04:05"),
	}}, *msgs...)
}

// AddWithAnime добавить сообщение с карточкой аниме
func (msgs *Messages) AddWithAnime(text string, animes models.Animes) {
	*msgs = append([]Message{{
		FromSystem: true,
		Message:    text,
		Time:       time.Now().Format("15:04:05"),
		NeedAnime:  len(animes) > 0,
		Animes:     animes,
	}}, *msgs...)
}
