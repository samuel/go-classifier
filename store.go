package classifier

import (
	"errors"
)

var (
	ErrCategoryAlreadyExists = errors.New("classifier: category already exists")
	ErrNegativeCount         = errors.New("classifier: count would go negative")
)

type Store interface {
	Categories() ([]string, error)
	AddCategory(name string) error
	AddDocument(category string, tokens []string) error
	RemoveDocument(category string, tokens []string) error
	DocumentCounts() (map[string]int64, error)                                             // category -> count
	TokenCounts(categories []string, tokens []string) (map[string]map[string]int64, error) // category -> token -> count
}

type localStore struct {
	categories     []string
	documentCounts map[string]int64            // category -> count
	tokenCounts    map[string]map[string]int64 // category -> token -> count
}

func NewLocalStore() Store {
	return &localStore{
		categories:     make([]string, 0),
		documentCounts: make(map[string]int64),
		tokenCounts:    make(map[string]map[string]int64),
	}
}

func (ls *localStore) Categories() ([]string, error) {
	return ls.categories, nil
}

func (ls *localStore) AddCategory(name string) error {
	if _, ok := ls.documentCounts[name]; ok {
		return nil
	}
	ls.categories = append(ls.categories, name)
	ls.documentCounts[name] = 0
	ls.tokenCounts[name] = make(map[string]int64)
	return nil
}

func (ls *localStore) AddDocument(category string, tokens []string) error {
	ls.documentCounts[category]++
	fc := ls.tokenCounts[category]
	for _, token := range tokens {
		fc[token]++
	}
	return nil
}

func (ls *localStore) RemoveDocument(category string, tokens []string) error {
	if ls.documentCounts[category] < 1 {
		return ErrNegativeCount
	}
	ls.documentCounts[category]--
	fc := ls.tokenCounts[category]
	for _, token := range tokens {
		if fc[token] < 1 {
			return ErrNegativeCount
		}
		fc[token]--
	}
	return nil
}

func (ls *localStore) DocumentCounts() (map[string]int64, error) {
	return ls.documentCounts, nil
}

func (ls *localStore) TokenCounts(categories []string, tokens []string) (map[string]map[string]int64, error) {
	if categories == nil {
		categories = ls.categories
	}
	counts := make(map[string]map[string]int64, len(categories))
	for _, cat := range categories {
		tc := ls.tokenCounts[cat]
		counts2 := make(map[string]int64, len(tokens))
		for _, t := range tokens {
			counts2[t] = tc[t]
		}
		counts[cat] = counts2
	}
	return counts, nil
}
