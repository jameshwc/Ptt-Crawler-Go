# Ptt-Crawler-Go

Implement Ptt Crawler in go with goroutine

**Status: Developing**

## Usage

```go
import (
	crawler "github.com/jameshwc/Ptt-Crawler-Go"
)
const (
	pages        = 5000
	numOfRoutine = 100
	storePath    = "dat/"
)

func main() {
	p := crawler.NewPTT(storePath, pages, numOfRoutine)
	p.CrawlBoard("Gossiping")
}
```