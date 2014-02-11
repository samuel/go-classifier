package classifier

import (
	"math"
	"sort"
)

// Default values for tuneables
const (
	DefaultUnknownTokenProbability = 0.5
	DefaultUnknowntokenStrength    = 0.45
	DefaultMaxDiscriminators       = 150
	DefaultMinProbabilityStrength  = 0.1
)

// BayesianClassifier is a naive Bayes classifier
type BayesianClassifier struct {
	store     Store
	tokenizer Tokenizer

	UnknownTokenProbability float64
	UnknownTokenStrength    float64
	MaxDiscriminators       int
	MinProbabilityStrength  float64 // distance from 0.5
}

// NewBayesianClassifier returns a new BayesianClassifier with default settings
func NewBayesianClassifier(store Store, tokenizer Tokenizer) (*BayesianClassifier, error) {
	return &BayesianClassifier{
		store:                   store,
		tokenizer:               tokenizer,
		UnknownTokenProbability: DefaultUnknownTokenProbability,
		UnknownTokenStrength:    DefaultUnknowntokenStrength,
		MaxDiscriminators:       DefaultMaxDiscriminators,
		MinProbabilityStrength:  DefaultMinProbabilityStrength,
	}, nil
}

// AddCategory creates a new category if it doesn't already exist
func (bc *BayesianClassifier) AddCategory(name string) error {
	return bc.store.AddCategory(name)
}

// AddDocument trains the classifier by feeding it text that's in the given category
func (bc *BayesianClassifier) AddDocument(category string, text string) error {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return err
	}
	return bc.store.AddDocument(category, tokens)
}

// RemoveDocument untrains the classifier
func (bc *BayesianClassifier) RemoveDocument(category string, text string) error {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return err
	}
	return bc.store.RemoveDocument(category, tokens)
}

// CategoryProbabilities returns P(document is in category | document)
func (bc *BayesianClassifier) CategoryProbabilities(text string, categories []string) (map[string]float64, error) {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return nil, err
	}
	probs, err := bc.filteredProbabilities(tokens, categories)
	if err != nil {
		return nil, err
	}

	// Fisher's method for combining probabilities
	catProb := make(map[string]float64, len(categories))
	for cat, pr := range probs {
		p := 0.5
		n := len(pr)
		if n > 0 {
			p = 1.0
			pExp := 0.0
			for _, tp := range pr {
				p *= tp
				if p < 1e-200 { // prevent underflow
					var e int
					p, e = math.Frexp(p)
					pExp += float64(e)
				}
			}
			// Compute the natural log of the product = sum of the logs:
			// ln(x * 2**i) = ln(x) + i * ln(2).
			p = -2.0 * (math.Log(p) + pExp*math.Ln2)
			p = invChi2(p, 2*n)
		}
		catProb[cat] = p
	}

	return catProb, nil
}

func (bc *BayesianClassifier) filteredProbabilities(tokens []string, categories []string) (map[string][]float64, error) {
	probs, err := bc.probabilities(tokens, categories)
	if err != nil {
		return nil, err
	}

	for cat, pr := range probs {
		if bc.MinProbabilityStrength > 0.0 {
			i := 0
			for i < len(pr) {
				dist := math.Abs(pr[i] - 0.5)
				if dist >= bc.MinProbabilityStrength {
					i++
				} else {
					if i == len(pr)-1 {
						pr = pr[:len(pr)-1]
						break
					}
					pr[i] = pr[len(pr)-1]
					pr = pr[:len(pr)-1]
				}
			}
		}

		if bc.MaxDiscriminators > 0 && len(pr) > bc.MaxDiscriminators {
			sort.Sort(sort.Float64Slice(pr))
			pr = pr[:bc.MaxDiscriminators]
		}

		probs[cat] = pr
	}

	return probs, nil
}

// return P(document is in category | document contains token)
func (bc *BayesianClassifier) probabilities(tokens []string, categories []string) (map[string][]float64, error) {
	tokenCounts, err := bc.store.TokenCounts(nil, tokens)
	if err != nil {
		return nil, err
	}
	docCounts, err := bc.store.Categories()
	if err != nil {
		return nil, err
	}

	prob := make(map[string][]float64, len(categories))
	for _, cat := range categories {
		prob[cat] = make([]float64, len(tokens))
	}
	unk := bc.UnknownTokenStrength * bc.UnknownTokenProbability
	for tokenI, t := range tokens {
		rSum := 0.0 // Sum(P(token|category)) for all categories
		nSum := int64(0)
		for cat, counts := range tokenCounts {
			n := float64(docCounts[cat])
			if n < 1 {
				n = 1
			}
			rSum += float64(counts[t]) / n
			nSum += counts[t]
		}

		for cat, pr := range prob {
			p := bc.UnknownTokenProbability
			if rSum > 0.0 {
				tc := tokenCounts[cat][t]
				dc := docCounts[cat]
				if dc < 1 {
					dc = 1
				}
				r := float64(tc) / float64(dc) // P(token|category)

				p = r / rSum
				if p > 1.0 {
					// There is an error in the counts (token count > doc count)
					// TODO: should this be an error?
					p = 1.0
				}

				// Weight
				p = (unk + float64(nSum)*p) / (bc.UnknownTokenStrength + float64(nSum))
			}
			pr[tokenI] = p
		}
	}

	return prob, nil
}
