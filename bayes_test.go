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
	bc.AddDocument("spam", "this is spam")
	bc.AddDocument("ham", "this isn't maybe")
	p, err := bc.CategoryProbability("this isn't", "spam")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", p)
}
