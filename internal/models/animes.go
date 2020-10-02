package models

type Animes []Anime

func (a Animes) Pointers() []*Anime {
	var animes = make([]*Anime, len(a))
	for i := range a {
		animes[i] = &a[i]
	}
	return animes
}

func (a Animes) Top(n int) Animes {
	if len(a) < n {
		return a
	}
	return a[:n]
}

func (a Animes) Len() int           { return len(a) }
func (a Animes) Less(i, j int) bool { return a[i].D < a[j].D }
func (a Animes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
