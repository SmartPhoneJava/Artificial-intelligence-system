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
