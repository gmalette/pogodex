package index

import (
	"testing"
)

func setupIndex() *index {
	i := NewIndex(new(NullStorage))
	i.AddDocument("1", "titi toto tutu")
	i.AddDocument("2", "hello world")
	i.AddDocument("3", "hello toto")
	i.WaitForIndexing()
	return i
}

func TestRetrieveFromWordQuery(t *testing.T) {
	i := setupIndex()

	query := new(wordQuery)
	query.word = "hello"

	docs := i.Query(query)

	if len(docs) != 2 {
		t.Errorf("Expected 2 results, got %d", len(docs))
	}

	if docs[0].externalId != "2" {
		t.Errorf("Expected first document to be 2, got %s", docs[0].externalId)
	}

	if docs[1].externalId != "3" {
		t.Errorf("Expected first document to be 3, got %s", docs[1].externalId)
	}
}

func TestRetrieveFromOrQuery(t *testing.T) {
	i := setupIndex()

	q1 := new(wordQuery)
	q1.word = "hello"

	q2 := new(wordQuery)
	q2.word = "titi"

	query := new(orQuery)
	query.left = q1
	query.right = q2

	docs := i.Query(query)

	if len(docs) != 3 {
		t.Errorf("Expected 3 results, got %d", len(docs))
	}
}
