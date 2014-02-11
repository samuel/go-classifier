package classifier

import (
	"fmt"
	"testing"
)

func TestBayesianClassifier(t *testing.T) {
	bc, err := NewBayesianClassifier(NewLocalStore(), SimpleTokenizer)
	if err != nil {
		t.Fatal(err)
	}
	bc.AddCategory("spam")
	bc.AddCategory("ham")
	bc.AddDocument("spam", "this is spam what")
	bc.AddDocument("ham", "this isn't maybe what")
	bc.AddDocument("ham", "what")
	text := "this what isn't"
	p, err := bc.CategoryProbabilities(text, []string{"spam", "ham"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", p)
}
