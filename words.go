package main

import (
	"strings"
)

type Word struct {
	word  string
	index int
}

func ProcessWords(rawWords []string) []Word {
	words := make([]Word, 0)
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}

func ProcessWordsFaster(rawWords []string) []Word {
	words := make([]Word, 0, len(rawWords))
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}

func process(word string) string { // simulate some sort of processing
	return strings.ToUpper(word)
}
