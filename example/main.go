package main

import (
	"fmt"

	crawler "github.com/jameshwc/Ptt-Crawler-Go"
)

func main() {
	p := crawler.NewPTT("test")
	fmt.Println(p.GetArticlesURL("Gossiping", 300))
	// if p.SetBoard("gossiping")

	// crawler.CrawlArticle("https://www.ptt.cc/bbs/Wine/M.1588519995.A.BA8.html")

}
