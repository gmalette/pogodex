package index

import (
	"math/big"
)

type Query interface {
	Ids(WordMap) *big.Int
}

type wordQuery struct {
	word string
}

func (w *wordQuery) Ids(m WordMap) *big.Int {
	ids := big.NewInt(0)
	word, ok := m.Get(w.word)

	if !ok {
		return ids
	}

	return ids.Or(ids, word.documents)
}

type orQuery struct {
	left  Query
	right Query
}

func (q *orQuery) Ids(m WordMap) *big.Int {
	ids := big.NewInt(0)
	ids.Or(q.left.Ids(m), q.right.Ids(m))
	return ids
}

type andQuery struct {
	right Query
	left  Query
}

func (q *andQuery) Ids(m WordMap) *big.Int {
	ids := big.NewInt(0)
	ids.And(q.left.Ids(m), q.right.Ids(m))
	return ids
}

type notQuery struct {
	query Query
}

func (q *notQuery) Ids(m WordMap) *big.Int {
	ids := big.NewInt(0)
	ids.SetBit(ids, int(m.Len()) + 1, 1)
	ids.Sub(ids, big.NewInt(1))
	ids.AndNot(ids, q.query.Ids(m))
	return ids
}
