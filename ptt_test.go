package ptt

import (
	"testing"
)

func Benchmark_CrawlBoardWithoutPrint(t *testing.B) {
	p := NewPTT(testDir, testPages, testNumOfRoutine, testPagePerFile)
	for i := 0; i < t.N; i++ {
		p.CrawlBoard("Gossiping")
	}
}
