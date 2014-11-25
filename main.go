package main

import (
	"fmt"
	"github.com/davecheney/profile"
	"io/ioutil"
	"os"
	"path/filepath"
	"pogodex/index"
	"runtime"
)

var max = 5000
var id = index.NewIndex()
var count = 1

func main() {
	runtime.GOMAXPROCS(4)
	cfg := profile.Config{
		CPUProfile:     true,
		ProfilePath:    ".",  // store profiles in current directory
		NoShutdownHook: true, // do not hook SIGINT
	}

	// p.Stop() must be called before the program exits to
	// ensure profiling information is written to disk.
	p := profile.Start(&cfg)
	defer p.Stop()
	seed()
}

func indexDocument(path string) {
	if count >= max {
		return
	}

	count++

	fileContent, _ := ioutil.ReadFile(path)
	str := string(fileContent)

	id.AddDocument(path, str)
}

func seed() {
	filepath.Walk("./data/", withFile)
	id.Stats()
	q := index.BuildQuery("")
	id.WaitForIndexing()
	docs := id.Query(q)
	fmt.Println(docs)
}

func withFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	fmt.Println(path)
	indexDocument(path)
	return nil
}
