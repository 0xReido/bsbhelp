package main

import (
	"log"
	"testing"
)

func BenchmarkGenerateMetadata(b *testing.B) {
	g, err := newGenerator("../../images")
	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := g.GenerateMetadata(n)
		if err != nil {
			b.Error(err)
		}
	}
}
