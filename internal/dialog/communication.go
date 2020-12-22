package dialog

type Communication struct {
	Messages   *Messages
	ToDialog   chan string
	FromDialog chan NLPResponse
}

func NewCommunication() *Communication {
	msgs := make(Messages, 0)
	msgs.Add("Привет, поболтаем? Ты можешь общаться со мной через текстовое поле над моей головой!", true)
	return &Communication{
		Messages:   &msgs,
		ToDialog:   make(chan string, 1),
		FromDialog: make(chan NLPResponse, 1),
	}
}

func (c Communication) Close() {
	close(c.ToDialog)
	close(c.FromDialog)
}
