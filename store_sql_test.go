// +build sqlite3

package classifier

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func createTables(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE categories (
        id INTEGER PRIMARY KEY ASC,
        name TEXT NOT NULL,
        document_count INTEGER NOT NULL DEFAULT 0,
        UNIQUE(name))`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE tokens (
        id INTEGER PRIMARY KEY ASC,
        category_id INTEGER NOT NULL,
        token TEXT NOT NULL,
        count INTEGER NOT NULL DEFAULT 0,
        FOREIGN KEY(category_id) REFERENCES categories(id),
        UNIQUE(category_id, token))`)
	return err
}

func TestSQLStore(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = createTables(db)
	if err != nil {
		t.Fatal(err)
	}

	store, err := NewSQLStore(db)
	if err != nil {
		t.Fatal(err)
	}

	if cats, err := store.Categories(); err != nil {
		t.Fatal(err)
	} else if len(cats) != 0 {
		t.Fatal("Number of categoreis should be 0")
	}

	if err := store.AddCategory("spam"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddCategory("ham"); err != nil {
		t.Fatal(err)
	}

	if cats, err := store.Categories(); err != nil {
		t.Fatal(err)
	} else if len(cats) != 2 {
		t.Fatal("Number of categoreis should be 2")
	} else if n, ok := cats["spam"]; n != 0 || !ok {
		t.Fatal("spam category not create properly")
	} else if n, ok := cats["ham"]; n != 0 || !ok {
		t.Fatal("ham category not create properly")
	}

	if err := store.AddDocument("none", []string{"blah"}); err != ErrCategoryDoesNotExist("none") {
		t.Fatalf("Expected ErrCategoryDoesNotExist not %+v", err)
	}

	if err := store.AddDocument("spam", []string{"this", "spam", "What"}); err != nil {
		t.Fatal(err)
	}

	if cats, err := store.Categories(); err != nil {
		t.Fatal(err)
	} else if n, ok := cats["spam"]; n != 1 || !ok {
		t.Fatalf("spam category expecetd n=1 && ok=true instead of n=%d && ok=%+v", n, ok)
	}

	if tokens, err := store.TokenCounts([]string{"spam", "ham"}, []string{"this"}); err != nil {
		t.Fatal(err)
	} else if toks, ok := tokens["spam"]; !ok {
		t.Fatal("Failed to find tokens for 'spam' category")
	} else if n := toks["this"]; n != 1 {
		t.Fatalf("Expected count for 'this' in category 'spam' to be 1 instead of %d", n)
	}

	if err := store.RemoveDocument("spam", []string{"this"}); err != nil {
		t.Fatal(err)
	}

	if cats, err := store.Categories(); err != nil {
		t.Fatal(err)
	} else if n, ok := cats["spam"]; n != 0 || !ok {
		t.Fatalf("spam category expecetd n=0 && ok=true instead of n=%d && ok=%+v", n, ok)
	}

	if tokens, err := store.TokenCounts([]string{"spam", "ham"}, []string{"this"}); err != nil {
		t.Fatal(err)
	} else if toks, ok := tokens["spam"]; !ok {
		t.Fatal("Failed to find tokens for 'spam' category")
	} else if n := toks["this"]; n != 0 {
		t.Fatalf("Expected count for 'this' in category 'spam' to be 0 instead of %d", n)
	}
}

func TestSQLStoreClassifier(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = createTables(db)
	if err != nil {
		t.Fatal(err)
	}

	store, err := NewSQLStore(db)
	if err != nil {
		t.Fatal(err)
	}

	bc, err := NewBayesianClassifier(store, SimpleTokenizer)
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
