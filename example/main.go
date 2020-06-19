package main

import (
	crawler "github.com/jameshwc/Ptt-Crawler-Go"
)

const (
	pages        = 1
	numOfRoutine = 100
	storePath    = "dat/"
)

func main() {
	p := crawler.NewPTT(storePath, pages, numOfRoutine)
	// go p.CrawlBoard("Gossiping")
	// go p.CrawlBoard("Womentalk")
	// for {
	// 	time.Sleep(10 * time.Second)
	// }
	p.CrawlBoard("Gossiping")
}
