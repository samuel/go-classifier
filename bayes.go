package classifier

import (
	"errors"
	"math"
	"sort"
)

var ErrStoreCountInvalid = errors.New("classifier: token in more documents than documents trained")

type tokenProbability struct {
	token string
	p     float64
}

type tokenProbabilityList []tokenProbability

func (l tokenProbabilityList) Len() int {
	return len(l)
}

func (l tokenProbabilityList) Less(a, b int) bool {
	return l[b].p < l[a].p
}

func (l tokenProbabilityList) Swap(a, b int) {
	l[a], l[b] = l[b], l[a]
}

type BayesianClassifier struct {
	store                   Store
	tokenizer               Tokenizer
	UnknownTokenProbability float64
	UnknownTokenStrength    float64
	MaxDiscriminators       int
	MinProbabilityStrength  float64 // distance from 0.5
}

func NewBayesianClassifier(store Store, tokenizer Tokenizer) (*BayesianClassifier, error) {
	return &BayesianClassifier{
		store:                   store,
		tokenizer:               tokenizer,
		UnknownTokenProbability: 0.5,
		UnknownTokenStrength:    0.45,
		MaxDiscriminators:       150,
		MinProbabilityStrength:  0.1,
	}, nil
}

func (bc *BayesianClassifier) AddCategory(name string) error {
	return bc.store.AddCategory(name)
}

func (bc *BayesianClassifier) AddDocument(category string, text string) error {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return err
	}
	return bc.store.AddDocument(category, tokens)
}

func (bc *BayesianClassifier) RemoveDocument(category string, text string) error {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return err
	}
	return bc.store.RemoveDocument(category, tokens)
}

func (bc *BayesianClassifier) CategoryProbability(text, category string) (float64, error) {
	tokens, err := bc.tokenizer.Tokenize(text)
	if err != nil {
		return 0.0, err
	}
	probs, err := bc.limitedProbabilities(tokens, category)
	if err != nil {
		return 0.0, err
	}

	// This is pretty much verbatim from spambayes/classifier.py
	p := 0.5
	n := len(probs)
	if n > 0 {
		h := 1.0
		s := 1.0
		hExp := 0.0
		sExp := 0.0

		for _, tp := range probs {
			s *= 1.0 - tp.p
			h *= tp.p
			if s < 1e-200 { // prevent underflow
				var e int
				s, e = math.Frexp(s)
				sExp += float64(e)
			}
			if h < 1e-200 { // prevent underflow
				var e int
				h, e = math.Frexp(h)
				hExp += float64(e)
			}
		}

		// Compute the natural log of the product = sum of the logs:
		// ln(x * 2**i) = ln(x) + i * ln(2).
		s = math.Log(s) + sExp*math.Ln2
		h = math.Log(h) + hExp*math.Ln2

		s = 1.0 - invChi2(-2.0*s, 2*n)
		h = 1.0 - invChi2(-2.0*h, 2*n)

		p = (s - h + 1.0) / 2.0
	}

	return p, nil
}

func (bc *BayesianClassifier) limitedProbabilities(tokens []string, category string) ([]tokenProbability, error) {
	probs, err := bc.probabilities(tokens, category)
	if err != nil {
		return nil, err
	}

	tProbs := make([]tokenProbability, 0, len(probs))
	for i, p := range probs {
		dist := math.Abs(p - 0.5)
		if dist >= bc.MinProbabilityStrength {
			t := tokens[i]
			tProbs = append(tProbs, tokenProbability{t, p})
		}
	}

	sort.Sort(tokenProbabilityList(tProbs))

	if len(tProbs) > bc.MaxDiscriminators {
		tProbs = tProbs[:bc.MaxDiscriminators]
	}

	return tProbs, nil
}

// return P(document is in category | document contains token)
func (bc *BayesianClassifier) probabilities(tokens []string, category string) ([]float64, error) {
	tokenCounts, err := bc.store.TokenCounts(nil, tokens)
	if err != nil {
		return nil, err
	}
	docCounts, err := bc.store.DocumentCounts()
	if err != nil {
		return nil, err
	}
	tc := tokenCounts[category]
	dc := docCounts[category]
	if dc == 0 {
		dc = 1
	}

	prob := make([]float64, len(tokens))
	unk := bc.UnknownTokenStrength * bc.UnknownTokenProbability
	for i, t := range tokens {
		n := tc[t]
		r := float64(n) / float64(dc) // P(token|category)

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

		p := bc.UnknownTokenProbability
		if rSum > 0.0 {
			p = r / rSum
			if p > 1.0 {
				// There is an error in the counts (token count > doc count)
				// TODO: should this be an error?
				p = 1.0
			}

			// Weight
			p = (unk + float64(nSum)*p) / (bc.UnknownTokenStrength + float64(nSum))
		}

		prob[i] = p
	}

	return prob, nil
}
