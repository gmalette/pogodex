package index

import (
	"testing"
	"math/big"
)

func setupWordMap() *index {
	i := NewIndex(new(NullStorage))

	words := new(wordHashMap)
	words.dict = make(map[string]*word)

	words.Add(&word{"hello", big.NewInt(1)})
	words.Add(&word{"world", big.NewInt(2)})
	words.Add(&word{"toto", big.NewInt(4)})
	words.Add(&word{"all", big.NewInt(7)})

	i.words = words

	return i
}

func newWordQuery(str string) *wordQuery {
	query := new(wordQuery)
	query.word = str
	return query
}

func TestWordQuery(t *testing.T) {
	index := setupWordMap()

	if str := newWordQuery("hello").Ids(index).String(); str != "1" {
		t.Errorf("Expected results for 'hello' to be 1, got %s", str)
	}

	if str := newWordQuery("world").Ids(index).String(); str != "2" {
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
	index := setupWordMap()

	helloQuery := newWordQuery("hello")
	worldQuery := newWordQuery("world")
	totoQuery := newWordQuery("toto")
	allQuery := newWordQuery("all")

	if str := newOrQuery(helloQuery, worldQuery).Ids(index).String(); str != "3" {
		t.Errorf("Expected results for 'hello' or 'world' to be 3, got %s", str)
	}

	doubleOrQuery := newOrQuery(newOrQuery(helloQuery, worldQuery), totoQuery)

	if str := doubleOrQuery.Ids(index).String(); str != "7" {
		t.Errorf("Expected results for 'hello' or 'world' or 'toto' to be 7, got %s", str)
	}

	if str := newOrQuery(allQuery, helloQuery).Ids(index).String(); str != "7" {
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
	index := setupWordMap()

	helloQuery := newWordQuery("hello")
	worldQuery := newWordQuery("world")
	totoQuery := newWordQuery("toto")
	allQuery := newWordQuery("all")

	noResultQuery := newAndQuery(helloQuery, worldQuery)

	if str := noResultQuery.Ids(index).String(); str != "0" {
		t.Errorf("Expected results for 'hello' and 'world' to be 0, got %s", str)
	}

	if str := newAndQuery(totoQuery, allQuery).Ids(index).String(); str != "4" {
		t.Errorf("Expected results for 'toto' and 'all' to be 4, got %s", str)
	}
}
