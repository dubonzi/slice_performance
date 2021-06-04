package main

import (
	"strings"
	"testing"
)

func BenchmarkProcessWords(b *testing.B) {
	words := strings.Split(book, " ")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ProcessWords(words)
	}
}

func BenchmarkProcessWordsFaster(b *testing.B) {
	words := strings.Split(book, " ")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ProcessWordsFaster(words)
	}
}
