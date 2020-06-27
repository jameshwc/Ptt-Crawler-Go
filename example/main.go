package main

import (
	crawler "github.com/jameshwc/Ptt-Crawler-Go"
)

const (
	pages        = 10000
	numOfRoutine = 32
	storePath    = "dat/"
	pagePerFile  = 100
)

func main() {
	p := crawler.NewPTT(storePath, pages, numOfRoutine, pagePerFile)
	// go p.CrawlBoard("Gossiping")
	// go p.CrawlBoard("Womentalk")
	// for {
	// 	time.Sleep(10 * time.Second)
	// }
	// p.CrawlURLlistToFile("Gossiping", 35000, 36000, "url.txt")
	p.CrawlBoard("Gossiping")
}
