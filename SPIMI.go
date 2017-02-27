package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Token struct {
	term  string
	docID uint16
}

var iter uint8
var size int64

func makeIndex(files []string) {
	iter = 0
	size = 0
	var blockSize uint16
	blockSize = 200
	filesNumber := len(files) / 2
	go setup(blockSize, files[filesNumber:])
	setup(blockSize, files[:filesNumber])
	mergeBlocks()
	fmt.Println("Collection size:", size)
}

func setup(blockSize uint16, files []string) {
	lenFiles := len(files)
	var i uint16
	for i = 0; i < uint16(lenFiles); i += blockSize {
		block := []*Token{}
		var docID uint16
		for docID = i; docID < i+blockSize; docID++ {
			if docID > uint16(lenFiles-1) {
				fmt.Println("return")
				return
			}
			fd, err := os.Open(files[docID])
			check(err)

			block = append(block, parseDocument(fd, docID)...)

		}
		invert(block)

		// debug.FreeOSMemory()
		log.Printf("%d / %d", i/blockSize, uint16(len(files))/blockSize)
	}
}

func parseDocument(r io.Reader, docID uint16) []*Token {
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanTerms)

	block := []*Token{}
	for scanner.Scan() {
		token := strings.ToLower(scanner.Text())
		block = append(block, &Token{token, docID})
		size++
	}
	return block
}

func invert(block []*Token) {
	blockFilename := "res/" + strconv.Itoa(int(iter))
	iter++
	dictionary := make(map[string][]uint16)
	for _, token := range block {
		postingList, ok := dictionary[token.term]
		if ok {
			postingList = append(postingList, token.docID)
		} else {
			dictionary[token.term] = []uint16{}
		}
	}
	serializeFile(blockFilename, dictionary)
}

func mergeBlocks() {
	blocksFiles := getFilesNames("res")
	dictionaryMerged := make(map[string][]uint16, 10000)
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
	fmt.Println("Dictionary size:", len(dictionaryMerged))
	serializeFile("dictionary", dictionaryMerged)
}

func removePunctuation(r rune) rune {
	if strings.ContainsRune(".,:;?!'&()`\"{}|[]#$%_*/\\><[]^`", r) {
		return -1
	}
	return r
}

func serializeFile(fileName string, dictionary map[string][]uint16) {
	f, err := os.Create(fileName)
	check(err)
	defer f.Close()

	dataEncoder := gob.NewEncoder(f)
	err2 := dataEncoder.Encode(dictionary)
	dictionary = nil
	check(err2)
}

func deserializeFile(fileName string) map[string][]uint16 {
	var data map[string][]uint16

	dataFile, err := os.Open(fileName)
	check(err)

	defer dataFile.Close()

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	return data
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ScanTerms(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var start int
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isControlBreak(r) {
			break
		}
	}

	for i := start; i < len(data); {
		r, width := utf8.DecodeRune(data[i:])
		if isControlBreak(r) {
			return i + width, data[start:i], nil
		}
		i += width
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	return start, nil, nil
}

func isPunctuation(r rune) bool {

	if r == '\'' || r == '-' {
		return false
	}
	if r >= '!' && r <= '/' ||
		r >= ':' && r <= '@' ||
		r >= '[' && r <= '`' ||
		r >= '{' && r <= '~' {
		return true
	}

	return false
}

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}
