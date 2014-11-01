package main

import (
	"./index"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
)

var max = 50000
var id = index.NewIndex(max + 1)
var count = 1

func main() {
	seed()
}

func indexDocument(path string) {
	if count >= max {
		return
	}

	count++

	fileContent, _ := ioutil.ReadFile(path)
	str := string(fileContent)

	id.AddDocument(str)
}

func seed() {
	filepath.Walk("./data/", withFile)
	id.Stats()
	q := index.BuildQuery("")
	ids := q.Ids(id)
	id.DocumentsByIds(ids)
}

func withFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	fmt.Println(path)
	indexDocument(path)
	return nil
}

