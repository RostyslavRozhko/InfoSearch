package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

type Token struct {
	term  string
	docID int
}

var iter int

func makeIndex(files []string) {
	iter = 0
	blockSize := 200
	//filesNumber := len(files)
	go setup(blockSize, 0, 200, files)
	setup(blockSize, 0, 200, files)
	//mergeBlocks()
}

func setup(blockSize int, start int, end int, files []string) {
	for i := start; i < end; i += blockSize {
		block := []*Token{}
		for docID := i; docID < i+blockSize; docID++ {
			fd, err := os.Open(files[docID])
			check(err)

			block = append(block, parseDocument(fd, docID)...)

		}
		invert(block)
		debug.FreeOSMemory()
		log.Printf("%d / %d", i/blockSize, len(files)/blockSize)
	}
}

func parseDocument(r io.Reader, docID int) []*Token {
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanTerms)

	block := []*Token{}
	for scanner.Scan() {
		token := strings.ToLower(scanner.Text())
		block = append(block, &Token{token, docID})
	}
	return block
}

func invert(block []*Token) {
	blockFilename := "res/" + strconv.Itoa(iter)
	iter++
	dictionary := make(map[string][]int)
	for _, token := range block {
		postingList, ok := dictionary[token.term]
		if ok {
			postingList = append(postingList, token.docID)
		} else {
			dictionary[token.term] = []int{}
		}
	}
	serializeFile(blockFilename, dictionary)
}

func mergeBlocks() map[string][]int {
	blocksFiles := getFilesNames("res")
	dictionaryMerged := make(map[string][]int, 10000)
	for index, fileName := range blocksFiles {
		for key, value := range deserializeFile(fileName) {
			postingList, ok := dictionaryMerged[key]
			if ok {
				postingList = append(postingList, value...)
			} else {
				dictionaryMerged[key] = value
			}
		}
		log.Printf("%d / %d", index, len(blocksFiles))
	}
	return dictionaryMerged
}

func removePunctuation(r rune) rune {
	if strings.ContainsRune(".,:;?!'&()`\"{}|[]#$%_*/\\><[]^`", r) {
		return -1
	}
	return r
}

func serializeFile(fileName string, dictionary map[string][]int) {
	f, err := os.Create(fileName)
	check(err)
	defer f.Close()

	dataEncoder := gob.NewEncoder(f)
	err2 := dataEncoder.Encode(dictionary)
	dictionary = nil
	check(err2)
}

func deserializeFile(fileName string) map[string][]int {
	var data map[string][]int

	dataFile, err := os.Open(fileName)
	check(err)

	defer dataFile.Close()

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)
	// check(err)

	return data
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
