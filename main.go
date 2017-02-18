package main

import (
	"io/ioutil"
	"log"
	"time"
)

func main() {

	log.Printf("Start")
	start := time.Now()
	makeIndex(getFilesNames("files"))

	elapsed := time.Since(start)
	log.Printf("Time spent %s", elapsed)
}

func getFilesNames(path string) []string {
	files, _ := ioutil.ReadDir(path)

	stringFiles := []string{}

	for _, file := range files {
		stringFiles = append(stringFiles, path+"/"+file.Name())
	}
	return stringFiles
}
