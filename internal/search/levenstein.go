package search

import (
	"log"

	"github.com/agnivade/levenshtein"
)

type LevenshteinDistance struct {
	Data []Pair
}

func NewLevenshteinDistance(text []string) LevenshteinDistance {
	var pairs = make([]Pair, len(text))
	for i, str := range text {
		pairs[i] = Pair{Index: i, Value: str}
	}
	return LevenshteinDistance{Data: pairs}
}

type Pair struct {
	Value string
	Index int
}

// Closest - найти самый похожий тайтл
func (ld LevenshteinDistance) Closest(searched string) Pair {
	var (
		minDistance    = -1
		minIndex       = -1
		minStringValue string
	)
	for _, text := range ld.Data {
		i := levenshtein.ComputeDistance(searched, text.Value)
		if i < minDistance || minDistance == -1 {
			minDistance = i
			minIndex = text.Index
			minStringValue = text.Value
		}
	}
	return Pair{minStringValue, minIndex}
}

// TyposLessN - получить список похожих, где отклонений меньше minDistance
func (ld LevenshteinDistance) TyposLessN(
	searched string,
	minDistance int,
) []Pair {
	return ld.FilterTyposLessN(ld.Data, searched, minDistance)
}

func (ld LevenshteinDistance) FilterTyposLessN(
	data []Pair,
	searched string,
	minDistance int,
) []Pair {
	var found = make([]Pair, 0, len(data))
	for _, text := range data {
		i := levenshtein.ComputeDistance(searched, text.Value)
		if i < minDistance {
			found = append(found, text)
		}
	}
	return found
}

// TyposLessP - получить список похожих, где отклонений меньше процента p от длины слова
func (ld LevenshteinDistance) TyposLessP(
	searched string,
	procent float32,
) []Pair {
	return ld.FilterTyposLessP(ld.Data, searched, procent)
}

func (ld LevenshteinDistance) FilterTyposLessP(
	data []Pair,
	searched string,
	procent float32,
) []Pair {
	var found = make([]Pair, 0, len(data))
	for _, text := range data {
		i := levenshtein.ComputeDistance(searched, text.Value)
		var proc = int(procent * float32(len(text.Value+searched)))
		//log.Println("before", i, proc, len(text.Value+searched), text)
		if i < proc {
			log.Println("after", float32(i)/float32(len(text.Value+searched)), procent, len(text.Value), text)
			found = append(found, text)
		}
	}
	return found
}

// ComputeDistance подcчитать расстояние Левенштейна
func (ld LevenshteinDistance) ComputeDistance(
	a, b string,
) int {
	return levenshtein.ComputeDistance(a, b)
}

func IntersectPairs(
	a, b []Pair,
) []Pair {
	var m = make(map[int]bool)
	for _, a := range a {
		m[a.Index] = true
	}
	var results = make([]Pair, 0)
	for _, b := range b {
		_, ok := m[b.Index]
		if ok {
			results = append(results, b)
		}
	}
	return results
}
