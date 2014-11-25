package index

import (
	"testing"
	"math/big"
)

func setupWordMap() WordMap {
	words := new(wordHashMap)
	words.dict = make(map[string]*word)

	words.Add(&word{"hello", big.NewInt(1)})
	words.Add(&word{"world", big.NewInt(2)})
	words.Add(&word{"toto", big.NewInt(4)})
	words.Add(&word{"all", big.NewInt(7)})

	return words
}

func newWordQuery(str string) *wordQuery {
	query := new(wordQuery)
	query.word = str
	return query
}

func TestWordQuery(t *testing.T) {
	words := setupWordMap()

	if str := newWordQuery("hello").Ids(words).String(); str != "1" {
		t.Errorf("Expected results for 'hello' to be 1, got %s", str)
	}

	if str := newWordQuery("world").Ids(words).String(); str != "2" {
		t.Errorf("Expected results for 'world' to be 2, got %s", str)
	}
}

func newOrQuery(left Query, right Query) *orQuery {
	query := new(orQuery)
	query.left = left
	query.right = right
	return query
}

func TestOrQuery(t *testing.T) {
	words := setupWordMap()

	helloQuery := newWordQuery("hello")
	worldQuery := newWordQuery("world")
	totoQuery := newWordQuery("toto")
	allQuery := newWordQuery("all")

	if str := newOrQuery(helloQuery, worldQuery).Ids(words).String(); str != "3" {
		t.Errorf("Expected results for 'hello' or 'world' to be 3, got %s", str)
	}

	doubleOrQuery := newOrQuery(newOrQuery(helloQuery, worldQuery), totoQuery)

	if str := doubleOrQuery.Ids(words).String(); str != "7" {
		t.Errorf("Expected results for 'hello' or 'world' or 'toto' to be 7, got %s", str)
	}

	if str := newOrQuery(allQuery, helloQuery).Ids(words).String(); str != "7" {
		t.Errorf("Expected results for 'all' or 'hello' to be 7, got %s", str)
	}
}

func newAndQuery(left Query, right Query) *andQuery {
	query := new(andQuery)
	query.left = left
	query.right = right
	return query
}

func TestAndQuery(t *testing.T) {
	words := setupWordMap()

	helloQuery := newWordQuery("hello")
	worldQuery := newWordQuery("world")
	totoQuery := newWordQuery("toto")
	allQuery := newWordQuery("all")

	noResultQuery := newAndQuery(helloQuery, worldQuery)

	if str := noResultQuery.Ids(words).String(); str != "0" {
		t.Errorf("Expected results for 'hello' and 'world' to be 0, got %s", str)
	}

	if str := newAndQuery(totoQuery, allQuery).Ids(words).String(); str != "4" {
		t.Errorf("Expected results for 'toto' and 'all' to be 4, got %s", str)
	}
}
