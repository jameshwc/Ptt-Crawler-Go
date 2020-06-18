package main

import (
	crawler "github.com/jameshwc/Ptt-Crawler-Go"
)

func main() {
	p := crawler.NewPTT()
	p.SetBoard("gossiping")
}
