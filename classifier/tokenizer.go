package classifier

import "strings"

// Tokenizer is the interface for a text tokenizer
type Tokenizer interface {
	Tokenize(text string) ([]string, error)
}

type simpleTokenizer struct{}

// SimpleTokenizer is a basic text tokenizer that splits on whitespace using strings.Fields,
// removes tokens shorter than 3 character, and removes duplicates.
var SimpleTokenizer = simpleTokenizer{}

func (t simpleTokenizer) Tokenize(text string) ([]string, error) {
	fields := strings.Fields(text)
	seen := make(map[string]bool)
	for i := 0; i < len(fields); {
		s := fields[i]
		fields[i] = strings.ToLower(s)
		if seen[s] || len(s) < 3 {
			if i == len(fields)-1 {
				fields = fields[:len(fields)-1]
				break
			}
			fields[i] = fields[len(fields)-1]
			fields = fields[:len(fields)-1]
		} else {
			seen[s] = true
			i++
		}
	}
	return fields, nil
}
