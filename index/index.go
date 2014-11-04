package index

import(
	"fmt"
	"strings"
	"math/big"
	"regexp"
	"sync/atomic"
)

type DocumentStorage interface {
	load(int32) (string, error)
	dump(int32, *string) error
}

type nullStorage struct {
}

func (n *nullStorage) load(id int32) (string, error) {
	return "", nil
}

func (n *nullStorage) dump(id int32, str *string) error {
	return nil
}

var Storage DocumentStorage = new(nullStorage)

type word struct {
	word string
	documents *big.Int
}

func (w* word) addDocument(doc *document) {
	w.documents.SetBit(w.documents, int(doc.id), 1)
}

type document struct {
	id int32
}

func NewDocument(id int32) *document {
	return &document{id}
}

type statistics struct {
}

type index struct {
	lastId int32
	defaultWordSize int
	words map[string]*word
	stats *statistics
	documents map[int]*document
	wordInsertLock chan int
	documentQueue chan *string
}

type Query interface {
	Ids(*index) *big.Int
}

type wordQuery struct {
	word string
}

func (w *wordQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	word, ok := i.words[w.word]

	if !ok {
		return ids
	}

	return ids.Or(ids, word.documents)
}

type orQuery struct {
	left Query
	right Query
}

func (q *orQuery) Ids(i *index) *big.Int {
	ids := big.NewInt(0)
	ids.Or(q.left.Ids(i), q.right.Ids(i))
	return ids
}

type andQuery struct {
	right Query
	left Query
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
	ids.SetBit(ids, len(i.documents) + 1, 1)
	ids.Sub(ids, big.NewInt(1))
	ids.AndNot(ids, q.query.Ids(i))
	return ids
}

func BuildQuery(query string) Query {
	aq := new(wordQuery)
	aq.word = "a"

	return aq
}

func NewIndex(size int) *index {
	index := new(index)
	index.words = make(map[string]*word)
	index.stats = new(statistics)
	index.lastId = 0
	index.defaultWordSize = size
	index.documents = make(map[int]*document)
	index.wordInsertLock = make(chan int, 1)
	index.wordInsertLock <- 1
	index.documentQueue = make(chan *string, 16)

	for i := 0; i < 5; i++ {
		go index.indexDocuments()
	}

	return index
}

func (i *index) indexDocuments() {
	for {
		content := <- i.documentQueue

		id := i.nextId()

		doc := new(document)
		doc.id = id

		Storage.dump(id, content)

		tokens := tokenize(content)

		missingWords := make([]string, len(tokens))
		missingWordCount := 0

		for _, word := range(tokens) {
			if _, ok := i.words[word]; !ok {
				missingWords[missingWordCount] = word
				missingWordCount++
			}
		}

		if missingWordCount > 0 {
			<- i.wordInsertLock
			fmt.Println("Missing word count: ", missingWordCount)
			for _, word := range missingWords[:missingWordCount] {
				i.words[word] = i.newWord(word)
			}
			i.wordInsertLock <- 1
		}

		for _, word := range(tokens) {
			w := i.words[word]
			w.addDocument(doc)
		}

		i.documents[int(id)] = doc
	}
}

func (i *index) DocumentsByIds(bitArray *big.Int) []*document {
	if bitArray.BitLen() == 0 {
		return make([]*document, 0)
	}

	docs := make([]*document, 0, bitArray.BitLen() + 1)

	for j := 0; j <= bitArray.BitLen(); j++ {
		if bitArray.Bit(j) != 0 {
			docs = docs[:len(docs) + 1]
			doc := i.documents[j]
			docs[len(docs) - 1] = doc
		}
	}

	fmt.Println("Results: ", docs)
	fmt.Println("Results Count", len(docs))
	return docs
}

func (i *index) nextId() int32 {
	return atomic.AddInt32(&i.lastId, 1)
}

func (i *index) newWord(str string) *word {
	documents := big.NewInt(0)
	w := new(word)
	w.documents = documents
	w.word = str
	return w
}

func (i *index) AddDocument(content string) {
	i.documentQueue <- &content
}

func (i *index) Stats() {
	fmt.Println("Words: ", len(i.words))
	fmt.Println("Documents: ", i.lastId)
}

func tokenize(content *string) []string {
	str := strings.ToLower(*content)
	rex := regexp.MustCompile("[[:word:]-_]+")

	matches := rex.FindAllStringSubmatch(str, -1)

	if matches == nil {
		return make([]string, 0)
	}

	tokens := make([]string, len(matches))

	for index, token := range(matches) {
		tokens[index] = token[0]
	}

	return tokens
}

