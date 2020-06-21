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
	if endPage == -1 {
		endPage = getLastArticlePage(board)
	}
	sem := make(chan int, p.numOfRoutine)
	n := endPage - startPage + 1
	pageList := make(chan []string, n)
	errc := make(chan error, n)
	if n < p.numOfRoutine {
		p.numOfRoutine = n
	}
	wg := new(sync.WaitGroup)
	if n%p.numOfRoutine == 0 {
		wg.Add(p.numOfRoutine)
	} else {
		wg.Add(p.numOfRoutine + 1)
	}
	for i, j := startPage, startPage+n/p.numOfRoutine; ; j += n / p.numOfRoutine {
		fmt.Println(i, j, n)
		sem <- 1
		if j >= endPage {
			go getArticleListThread(p.baseURL, board, i, endPage, sem, pageList, errc, wg)
			break
		}
		go getArticleListThread(p.baseURL, board, i, j, sem, pageList, errc, wg)
		time.Sleep(50)
		i = j + 1
	}
	wg.Wait()
	close(pageList)
	close(errc)
	for i := range pageList {
		URLs = append(URLs, i...)
	}
	if len(errc) != 0 {
		e = <-errc
	} else {
		e = nil
	}
	return
}

func (p *PTT) GetArticlesURL(board string, start, end int) (URLs []string, e error) {
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

func getArticleListThread(baseURL, board string, start, end int, sem chan int, list chan []string, e chan error, wg *sync.WaitGroup) {
	endFunc := func() {
		<-sem
		wg.Done()
	}
	defer endFunc()
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
