package ptt

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (p *PTT) getArticlesURLThread(board string, startPage, endPage int) (URLs []string, e error) {
	if endPage < 0 {
		endPage = getLastArticlePage(board)
	}
	n := endPage - startPage + 1
	pageList := make(chan []string, n)
	errc := make(chan error, n)
	if n < p.numOfRoutine {
		p.numOfRoutine = n
	}
	wg := new(sync.WaitGroup)
	if n%p.numOfRoutine == 0 {
		wg.Add(p.numOfRoutine - 1)
	} else {
		wg.Add(p.numOfRoutine)
	}
	counter := 0
	for i, j := startPage, startPage+n/p.numOfRoutine; ; j += n / p.numOfRoutine {
		if j >= endPage {
			go getArticleListThread(p.baseURL, board, i, endPage, pageList, errc, wg)
			break
		}
		go getArticleListThread(p.baseURL, board, i, j, pageList, errc, wg)
		time.Sleep(100)
		i = j + 1
		counter++
	}
	wg.Wait()
	close(pageList)
	close(errc)
	for i := range pageList {
		URLs = append(URLs, i...)
	}
	if len(errc) != 0 {
		log.Printf("Has %d Errors!\n", len(errc))
		e = <-errc
	} else {
		e = nil
	}
	fmt.Printf("Completely downloading URLlist...Got %d articles ready to download...\n", len(URLs))
	return
}

func (p *PTT) getArticlesURL(board string, start, end int) (URLs []string, e error) {
	return getArticleList(p.baseURL, board, start, end)
}
func getArticleList(baseURL, board string, start, end int) ([]string, error) {
	var articleList []string
	for i := start; i <= end; i++ {
		doc, err := parseURL(baseURL + "bbs/" + board + "/index" + strconv.Itoa(i) + ".html")
		if err != nil {
			return articleList, err
		}
		doc.Find(".title").Each(func(i int, s *goquery.Selection) {
			val, exist := s.Children().Attr("href")
			if exist {
				articleList = append(articleList, baseURL+val[1:])
			}

		})
	}
	return articleList, nil
}
func getLastArticlePage(url string) int {
	left, right := 0, 100000
	for left+1 < right {
		mid := (left + right) / 2
		if checkArticlePage(url, mid) {
			left = mid
		} else {
			right = mid
		}
	}
	return left - 1 // To avoid overlapped articles
}

func getArticleListThread(baseURL, board string, start, end int, list chan []string, e chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if start > end {
		e <- fmt.Errorf("getArticleList: start %d is greater than end %d", start, end)
		return
	}
	var articleList []string
	for i := start; i <= end; i++ {
		doc, err := parseURL(baseURL + "bbs/" + board + "/index" + strconv.Itoa(i) + ".html")
		if err != nil {
			e <- err
			return
		}
		doc.Find(".title").Each(func(i int, s *goquery.Selection) {
			val, exist := s.Children().Attr("href")
			if exist {
				articleList = append(articleList, baseURL+val[1:])
			}
		})
	}
	list <- articleList
}
func checkArticlePage(url string, n int) bool {
	url = url + strconv.Itoa(n) + ".html"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("checkArticlePage: ", err)
		return false
	}
	req.AddCookie(over18cookie)
	resp, err := defaultClient.Do(req)
	if err != nil {
		log.Println("checkArticlePage: ", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
