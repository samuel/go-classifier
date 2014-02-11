package classifier

import (
	"errors"
)

var (
	// ErrNegativeCount is the error that's returned when a count goes or is about to go negative.
	ErrNegativeCount = errors.New("classifier: count would go negative")
)

// ErrCategoryDoesNotExist is the error returned when a category doesn't exist.
type ErrCategoryDoesNotExist string

func (e ErrCategoryDoesNotExist) Error() string {
	return "classifier: category " + string(e) + " does not exist"
}

// Store is the storage interface for a classifier
type Store interface {
	Categories() (map[string]int64, error) // category -> document count
	AddCategory(name string) error
	AddDocument(category string, tokens []string) error
	RemoveDocument(category string, tokens []string) error
	TokenCounts(categories, tokens []string) (map[string]map[string]int64, error) // category -> token -> count
}

type localStore struct {
	categories     []string
	documentCounts map[string]int64            // category -> count
	tokenCounts    map[string]map[string]int64 // category -> token -> count
}

// NewLocalStore returns a new in-memory store (for testing purposes mainly)
func NewLocalStore() Store {
	return &localStore{
		categories:     make([]string, 0),
		documentCounts: make(map[string]int64),
		tokenCounts:    make(map[string]map[string]int64),
	}
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

func (ls *localStore) Categories() (map[string]int64, error) {
	return ls.documentCounts, nil
}

func (ls *localStore) TokenCounts(categories, tokens []string) (map[string]map[string]int64, error) {
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
