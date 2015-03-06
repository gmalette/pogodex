package index

import (
	"math/big"
)

type Query interface {
	Ids(*index) *big.Int
}

type wordQuery struct {
	word string
}

func (w *wordQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	word, ok := i.words.Get(w.word)

	if !ok {
		return ids
	}

	return ids.Or(ids, word.documents)
}

type orQuery struct {
	left  Query
	right Query
}

func (q *orQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	ids.Or(q.left.Ids(i), q.right.Ids(i))
	return ids
}

type andQuery struct {
	right Query
	left  Query
}

func (q *andQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	ids.And(q.left.Ids(i), q.right.Ids(i))
	return ids
}

type notQuery struct {
	query Query
}

func (q *notQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	ids.SetBit(ids, int(i.words.Len()) + 1, 1)
	ids.Sub(ids, big.NewInt(1))
	ids.AndNot(ids, q.query.Ids(i))
	return ids
}

type exactQuery struct {
	tokens []string
}

func (q *exactQuery) Ids(i *index) *big.Int {
	andQuery := q.buildQuery(q.tokens)

	superset := andQuery.Ids(i)
	return superset
}

func (q *exactQuery) buildQuery(tokens []string) Query {
	if len(tokens) == 0 {
		return &wordQuery{tokens[0]}
	}

	aq := new(andQuery)
	aq.left = &wordQuery{tokens[0]}
	aq.right = q.buildQuery(tokens[1:len(tokens) - 1])

	return aq
}
