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
	if err := bc.AddCategory("spam"); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddCategory("ham"); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddDocument("spam", "this is spam what"); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddDocument("ham", "this isn't maybe what"); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddDocument("ham", "what"); err != nil {
		t.Fatal(err)
	}
	text := "this what isn't"
	p, err := bc.CategoryProbabilities(text, []string{"spam", "ham"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", p)
}
