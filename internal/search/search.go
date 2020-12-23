package search

import (
	"log"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/sahilm/fuzzy"
	"github.com/schollz/closestmatch/levenshtein"
)

// Engine - ядро поиска
type Engine struct {
	Config              Config
	savedData           fuzzy.Source
	closestMatch        *levenshtein.ClosestMatch
	closestMatchToIndex map[string]int
}

// Stemming настройки стемминга
type Stemming struct {
	On       bool   // требуется убирать окончания
	Language string // язык стемминга
}

// Typos настройки опечаток
type Typos struct {
	On       bool // искать наиболее похожие слова, даже если в них другие символы
	BagSizes int  // кол-во ответов
}

// Config - настройки применения алгоритмов расширения результатов поиска
type Config struct {
	Stemming Stemming // требуется убирать окончания
	//SkipLetters bool   // найденный результат может содержать не все искомые буквы
	WithTypos Typos // если точных совпадений не найдено, то искать слова с опечатками
}

func NewSearchEngine(config Config, data fuzzy.Source) Engine {
	if config.Stemming.On && config.Stemming.Language == "" {
		config.Stemming.Language = "russian"
	}
	if config.WithTypos.On && config.WithTypos.BagSizes == 0 {
		config.WithTypos.BagSizes = 1
	}
	var engine = Engine{
		Config: config,
	}
	engine.initClosestMatch(data)

	return engine
}

func (engine *Engine) initClosestMatch(data fuzzy.Source) {
	engine.closestMatchToIndex = make(map[string]int, data.Len())
	if data != nil && data.Len() > 0 {
		engine.savedData = data
		if engine.Config.WithTypos.On {
			var (
				//bagSizes = []int{engine.Config.WithTypos.BagSizes}
				words = make([]string, data.Len())
			)
			for i := 0; i < data.Len(); i++ {
				str := data.String(i)
				words[i] = str
				engine.closestMatchToIndex[str] = i
			}
			log.Println("learn with", len(words))
			engine.closestMatch = levenshtein.New(words)
		}
	}
}

func (engine Engine) dataToPairs(
	data fuzzy.Source,
) []Pair {
	var strs = make([]Pair, data.Len())
	for i := 0; i < data.Len(); i++ {
		strs[i] = Pair{data.String(i), i}
	}
	return strs
}

func (engine Engine) fuzzyToPairs(
	matches fuzzy.Matches,
) []Pair {
	var strs = make([]Pair, len(matches))
	for i, match := range matches {
		strs[i] = Pair{match.Str, match.Index}
	}
	return strs
}

// Search запустить поиск
func (engine Engine) Search(
	name string,
	data fuzzy.Source,
) ([]Pair, error) {
	if name == "" || data == nil || data.Len() == 0 {
		return nil, nil
	}

	originName := name
	if engine.Config.Stemming.On {
		words := strings.Split(name, " ")
		for i, word := range words {
			stemmed, err := snowball.Stem(
				word,
				engine.Config.Stemming.Language,
				true,
			)
			if err != nil {
				return nil, err
			}
			words[i] = stemmed
		}
		name = strings.Join(words, " ")
	}

	log.Println(name)
	results := fuzzy.FindFrom(name, data)

	var (
		allPairs   = engine.dataToPairs(data)
		foundPairs = engine.fuzzyToPairs(results)

		levDist = LevenshteinDistance{allPairs}
	)

	// for _, v := range results {
	// 	log.Println("results1", v.Index, v.Str)
	// }

	//foundPairs = IntersectPairs(foundPairs, levDist.TyposLessP(originName, 0.3))

	//foundPairs = levDist.FilterTyposLessP(foundPairs, originName, 0.3)
	//log.Println("results1", len(foundPairs))

	allOk := len(IntersectPairs(foundPairs, levDist.TyposLessN(originName, len(originName)))) != 0

	if len(foundPairs) == 0 || len(foundPairs) > 10 || !allOk {
		foundPairs = levDist.TyposLessN(originName, 5)
	}
	// log.Println("results2", len(foundPairs))

	// for _, v := range foundPairs {
	// 	proc := float32(levDist.ComputeDistance(v.Value, name)) / float32(len(v.Value+name))
	// 	log.Println("results2", v.Index, v.Value, levDist.ComputeDistance(v.Value, name), proc)
	// }

	return foundPairs, nil
}
