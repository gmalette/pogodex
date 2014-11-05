package main

import (
	"pogodex/index"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"github.com/davecheney/profile"
	"runtime"
)

var max = 50000
var id = index.NewIndex(max + 1)
var count = 1

func main() {
runtime.GOMAXPROCS(4)
	cfg := profile.Config{
		CPUProfile: true,
		ProfilePath: ".",  // store profiles in current directory
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

	id.AddDocument(str)
}

func seed() {
	filepath.Walk("./data/", withFile)
	id.Stats()
	q := index.BuildQuery("")
	id.WaitForIndexing()
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

