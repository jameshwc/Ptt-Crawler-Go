package ptt

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (p *PTT) GetArticlesURLThread(board string, pages int) (URLs []string, e error) {
	sem := make(chan int, p.numOfRoutine)
	pageList := make(chan []string, pages)
	errc := make(chan error, pages)
	n := getLastArticlePage(p.baseURL + "bbs/" + board + "/index")
	counter := 0
	if pages < p.numOfRoutine {
		p.numOfRoutine = pages
	}
	for i, j := n-pages, n-pages+pages/p.numOfRoutine; ; j += pages / p.numOfRoutine {
		sem <- 1
		if j >= n {
			go getArticleListThread(p.baseURL, board, i, n, sem, pageList, errc)
			counter++
			break
		}
		go getArticleListThread(p.baseURL, board, i, j, sem, pageList, errc)
		time.Sleep(50)
		counter++
		i = j + 1
	}
	for len(pageList)+len(errc) != counter {
	}
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

func (p *PTT) GetArticlesURL(board string, pages int) (URLs []string, e error) {
	n := getLastArticlePage(p.baseURL + "bbs/" + board + "/index")
	return getArticleList(p.baseURL, board, n-pages, n)
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

func getArticleListThread(baseURL, board string, start, end int, sem chan int, list chan []string, e chan error) {
	endFunc := func() { <-sem }
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
