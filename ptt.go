package ptt

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
)

type PTT struct {
	baseURL          string
	bbsURL           string
	storePath        string
	numOfRoutine     int
	writeFileRoutine int
	pages            int
	pagePerFile      int
	delayTime        time.Duration
}

var (
	over18cookie  *http.Cookie = &http.Cookie{Name: "over18", Value: "1"}
	defaultClient *http.Client = &http.Client{}
	secondClient  *http.Client
)

func NewPTT(storePathFolder string, pages, numsOfRoutine, pagePerFile int, delayTime int64) *PTT {
	p := new(PTT)
	p.baseURL = "https://www.ptt.cc/"
	p.bbsURL = "https://www.ptt.cc/bbs/"
	p.storePath = storePathFolder
	p.numOfRoutine = numsOfRoutine
	p.pages = pages
	p.pagePerFile = pagePerFile
	p.delayTime = delayTime
	p.writeFileRoutine = 1
	proxyURL, err := url.Parse("http://205.185.115.100:8080")
	if err != nil {
		log.Fatal("Error when set up proxy: ", err)
	}
	secondClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	return p
}

func isValidBoard(bbsUrl, board string) bool {
	if resp, err := http.Get(bbsUrl + board + "/index.html"); err != nil {
		fmt.Println(err)
		return false
	} else if resp.StatusCode != 200 {
		return false
	}
	return true
}

func (p *PTT) CrawlBoard(board string) {
	if !isValidBoard(p.bbsURL, board) {
		log.Fatal("Boardname not valid!")
	}
	endPage := getLastArticlePage(p.getBoardURL(board))
	fmt.Printf("Now begin downloading from %d page to %d page...\n", endPage-p.pages, endPage)
	p.crawlBoard(board, endPage-p.pages, endPage)
}

func (p *PTT) CrawlBoardWithPages(board string, startPage, endPage int) {
	if !isValidBoard(p.bbsURL, board) {
		log.Fatal("Boardname not valid!")
	}
	latestPage := getLastArticlePage(p.getBoardURL(board))
	if latestPage+1 == endPage {
		endPage--
		log.Println("For performance and to avoid overlapped articles, we don't support downloading latest page. The end page has been changed to ", endPage)
	}
	if latestPage < endPage || startPage < 0 || startPage > endPage {
		log.Fatal("The Scope of pages is not correct!")
	}
	p.crawlBoard(board, startPage, endPage)
}

func (p *PTT) CrawlWithURLFile(inputFile, outputFile string) {
	URLlist, err := parseURLfile(filepath.Join(p.storePath, inputFile))
	if err != nil {
		log.Fatal("parseURLfile: ", err)
	}
	articles := p.crawlArticles(URLlist)
	saveFile(articles, outputFile)
}

func (p *PTT) CrawlURLlistToFile(board string, startPage, endPage int, filename string) {
	file, err := os.Create(filepath.Join(p.storePath, filename))
	if err != nil {
		log.Fatal(errors.New("SaveURLlistToFile: " + err.Error()))
	}
	defer file.Close()
	if latestPage := getLastArticlePage(p.getBoardURL(board)); endPage < 0 || endPage > latestPage {
		log.Fatalf("There's something wrong with endPage %d!", endPage)
	}
	URLlist, err := p.getArticlesURLThread(board, startPage, endPage)
	if err != nil {
		log.Fatal(errors.New("SaveURLlistToFile: " + err.Error()))
	}
	for i := range URLlist {
		file.WriteString(URLlist[i] + "\n")
	}
}

func (p *PTT) crawlBoard(board string, startPage, endPage int) {
	// p.crawlPages(board, startPage, endPage)
	start, end := startPage, startPage+p.pagePerFile-1
	// wg := new(sync.WaitGroup)
	// sem := make(chan int, p.writeFileRoutine)
	for start <= endPage {
		// 	//fmt.Println(start, end)
		// 	wg.Add(1)
		if end > endPage {
			end = endPage
		}
		// 	sem <- 1
		p.crawlPages(board, start, end)
		start = end + 1
		end += p.pagePerFile
		time.Sleep(p.delayTime)
	}
	// wg.Wait()
}
func (p *PTT) crawlPages(board string, start, end int) {
	URLlist, err := p.getArticlesURLThread(board, start, end)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Crawl Pages from %d to %d...\n", start, end)
	articles := p.crawlArticles(URLlist, fmt.Sprintf("Page %d-%d", start, end))
	filename := filepath.Join(p.storePath, board+"_P"+strconv.Itoa(start)+"_"+strconv.Itoa(end)+"_T"+time.Now().Format("0102_15")+".json")
	saveFile(articles, filename)
	// wg.Done()
	// <-sem
}
func (p *PTT) crawlArticles(URLlist []string, s ...interface{}) chan article {
	sem := make(chan int, p.numOfRoutine)
	n := len(URLlist)
	articles := make(chan article, n) // make sure channel has enough space if page == 1
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for i := range URLlist {
		sem <- 1
		go CrawlArticleThread(URLlist[i], articles, sem, wg)
		fmt.Printf("\r%s [%d/%d] %s", s, i+1, n, URLlist[i])
		time.Sleep(p.delayTime)
	}
	wg.Wait()
	close(articles)
	fmt.Println(" All articles downloaded!")
	return articles
}
func saveFile(articles chan article, filename string) {
	articleSlice := make([]article, len(articles))
	idx := 0
	for a := range articles {
		articleSlice[idx] = a
		idx++
	}
	jsonData, err := json.Marshal(articleSlice)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("Now create json file...")
	outputFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	outputFile.Write(jsonData)
}

func parseURLfile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("parseURLfile: " + err.Error())
	}
	scanner := bufio.NewScanner(file)
	set := mapset.NewSet()
	var URLlist []string
	for scanner.Scan() {
		URLlist = append(URLlist, scanner.Text())
		set.Add(scanner.Text())
	}
	fmt.Println(set.Cardinality(), len(URLlist))
	return URLlist, nil
}
func (p *PTT) getBoardURL(board string) string { // without .html
	return p.bbsURL + board + "/index"
}
