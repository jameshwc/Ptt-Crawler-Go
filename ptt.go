package ptt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
)

const (
	maxNumArticles = 15000
)

type PTT struct {
	baseURL      string
	bbsURL       string
	storePath    string
	numOfRoutine int
	pages        int
	board        mapset.Set
}

var (
	over18cookie  *http.Cookie = &http.Cookie{Name: "over18", Value: "1"}
	defaultClient *http.Client = &http.Client{}
)

func NewPTT(storePathFolder string, pages, numsOfRoutine int) *PTT {
	p := new(PTT)
	p.baseURL = "https://www.ptt.cc/"
	p.bbsURL = "https://www.ptt.cc/bbs/"
	p.storePath = storePathFolder
	p.numOfRoutine = numsOfRoutine
	p.board = mapset.NewSet()
	p.pages = pages
	return p
}

func (p *PTT) SetBoard(board string) error {
	if isValidBoard(p.bbsURL, board) {
		p.board.Add(board)
	} else {
		return fmt.Errorf("board name %s not valid", board)
	}
	return nil
}

func (p *PTT) SetBoardWithSlice(board []string) error {
	errMsg := ""
	for id := range board {
		if isValidBoard(p.bbsURL, board[id]) {
			p.board.Add(board[id])
		} else {
			errMsg += board[id] + " "
		}
	}
	if len(errMsg) > 0 {
		return errors.New("board name " + errMsg + "not valid")
	}
	return nil
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
	endPage := getLastArticlePage(p.bbsURL + board + "/index")
	startPage := endPage - p.pages
	URLlist, err := p.getArticlesURLThread(board, startPage, endPage)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Completely downloading URLlist...Got %d articles ready to download...\n", len(URLlist))
	articles := p.crawlArticles(URLlist)
	fmt.Println("All articles downloaded!")
	filename := filepath.Join(p.storePath, board+"_P"+strconv.Itoa(startPage)+"_"+strconv.Itoa(endPage)+"_T"+time.Now().Format("0102_15")+".json")
	saveFile(articles, filename)
}

func (p *PTT) crawlArticles(URLlist []string) chan article {
	sem := make(chan int, p.numOfRoutine)
	articles := make(chan article, len(URLlist)) // make sure channel has enough space if page == 1
	wg := new(sync.WaitGroup)
	wg.Add(len(URLlist))
	for i := range URLlist {
		sem <- 1
		go CrawlArticleThread(URLlist[i], articles, sem, wg)
		fmt.Println(i, URLlist[i])
		time.Sleep(100)
	}
	wg.Wait()
	close(articles)
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
	fmt.Println("Now create json file...")
	outputFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	outputFile.Write(jsonData)
}

func (p *PTT) getBoardURL(board string) string { // without .html
	return p.bbsURL + board + "/index"
}
