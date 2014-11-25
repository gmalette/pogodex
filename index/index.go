package index

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"github.com/quipo/statsd"
)

type DocumentStorage interface {
	load(int32) (string, error)
	dump(int32, *string) error
}

type nullStorage struct {
}

type WordMap interface {
	Get(string) (*word, bool)
	Add(*word)
	Len() uint
}

type wordHashMap struct {
	dict map[string]*word
}

func NewWordHashMap() *wordHashMap {
	m := new(wordHashMap)
	m.dict = make(map[string]*word)
	return m
}

func (w *wordHashMap) Len() uint {
	return uint(len(w.dict))
}

func (w *wordHashMap) Add(word *word) {
	w.dict[word.word] = word
}

func (w *wordHashMap) Get(key string) (*word, bool) {
	word, ok := w.dict[key]
	return word, ok
}

func (n *nullStorage) load(id int32) (string, error) {
	return "", nil
}

func (n *nullStorage) dump(id int32, str *string) error {
	return nil
}

var Storage DocumentStorage = new(nullStorage)

type word struct {
	word      string
	documents *big.Int
}

func (w *word) addDocument(doc *document) {
	w.documents.SetBit(w.documents, int(doc.id), 1)
}

type document struct {
	externalId string
	id int32
}

type statistics struct {
}

type documentContent struct {
	externalId string
	tokens *[]string
	content *string
}

type index struct {
	nextId          int32
	words           WordMap
	stats           *statistics
	documents       map[int]*document
	wordInsertLock  sync.Mutex
	tokenizeQueue   chan *documentContent
	indexQueue      chan *documentContent
	indexWait       sync.WaitGroup
	statsdClient    *statsd.StatsdClient
}

func BuildQuery(query string) Query {
	aq := new(wordQuery)
	aq.word = "a"

	return aq
}

func NewIndex() *index {
	index := new(index)
	index.words = NewWordHashMap()
	index.stats = new(statistics)
	index.nextId = 0
	index.documents = make(map[int]*document)
	index.tokenizeQueue = make(chan *documentContent, 16)
	index.indexQueue = make(chan *documentContent, 16)

	statsdClient := statsd.NewStatsdClient("localhost:8125", "pogodex.")
	statsdClient.CreateSocket()
	index.statsdClient = statsdClient

	index.statsdClient.Gauge("document_count", 1)

	for i := 0; i < 5; i++ {
		go index.tokenizeDocuments()
	}

	go index.indexDocuments()

	return index
}

func (i *index) Query(q Query) []*document {
	return i.DocumentsByIds(q.Ids(i.words))
}

func (i *index) indexDocuments() {
	for {
		docContent := <-i.indexQueue

		id := i.nextId
		i.nextId++

		doc := new(document)
		doc.externalId = docContent.externalId
		doc.id = id

		Storage.dump(id, docContent.content)

		tokens := *docContent.tokens

		for _, token := range tokens {
			word, ok := i.words.Get(token);
			if !ok {
				word = i.newWord(token)
				i.words.Add(word)
			}
			word.addDocument(doc)
		}

		i.documents[int(id)] = doc
		i.indexWait.Done()
	}
}

func (i *index) tokenizeDocuments() {
	for {
		docContent := <-i.tokenizeQueue
		docContent.tokens = tokenize(docContent.content)
		i.indexQueue <- docContent
	}
}

func (i *index) DocumentsByIds(bitArray *big.Int) []*document {
	if bitArray.BitLen() == 0 {
		return make([]*document, 0)
	}

	docs := make([]*document, 0, bitArray.BitLen()+1)

	for j := 0; j <= bitArray.BitLen(); j++ {
		if bitArray.Bit(j) != 0 {
			docs = docs[:len(docs)+1]
			doc := i.documents[j]
			docs[len(docs)-1] = doc
		}
	}

	fmt.Println("Results: ", docs)
	fmt.Println("Results Count", len(docs))
	return docs
}

func (i *index) newWord(str string) *word {
	documents := big.NewInt(0)
	w := new(word)
	w.documents = documents
	w.word = str
	return w
}

func (i *index) AddDocument(externalId string, content string) {
	i.indexWait.Add(1)
	docContent := documentContent{
		externalId,
		nil,
		&content,
	}
	i.tokenizeQueue <- &docContent
}

func (i *index) Stats() {
	fmt.Println("Words: ", i.words.Len())
	fmt.Println("Documents: ", i.nextId)
}

func (i *index) WaitForIndexing() {
	i.indexWait.Wait()
}

func uniqueWords(words *[]string) []string {
	length := len(*words) - 1
	unique := make([]string, len(*words))
	copy(unique, *words)

	for i := 0; i < length; i++ {
		for j := i + 1; j <= length; j++ {
			if unique[i] == unique[j] {
				unique[j] = unique[length]
				length--
				unique = unique[0:length]
				j--
			}
		}
	}

	return unique
}

func tokenize(content *string) *[]string {
	str := strings.ToLower(*content)
	rex := regexp.MustCompile("[[:word:]-_]+")

	matches := rex.FindAllStringSubmatch(str, -1)

	if matches == nil {
		var ret *[]string = new([]string)
		return ret
	}

	tokens := make([]string, len(matches))

	for index, token := range matches {
		tokens[index] = token[0]
	}

	return &tokens
}
