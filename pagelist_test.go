package ptt

import "testing"

func Benchmark_GetArticlesURLThread(b *testing.B) {
	p := NewPTT("")
	for i := 0; i < b.N; i++ {
		p.GetArticlesURLThread("Seniorhigh", 500)
	}
}

func Benchmark_GetArticlesURL(b *testing.B) {
	p := NewPTT("")
	for i := 0; i < b.N; i++ {
		p.GetArticlesURL("Seniorhigh", 500)
	}
}
