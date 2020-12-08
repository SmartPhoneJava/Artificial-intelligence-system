package dialog

import "time"

type Message struct {
	FromSystem bool
	Message    string
	Time       string
}

type Messages []Message

func (msgs *Messages) Add(text string, system bool) {
	*msgs = append([]Message{{
		FromSystem: system,
		Message:    text,
		Time:       time.Now().Format("15:04:05"),
	}}, *msgs...)
}
