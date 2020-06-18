package ptt

import (
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func (p *PTT) GetArticlesURL(board string, pages int) ([]string, error) {
	n := getLastArticlePage(p.baseURL + "bbs/" + board + "/index")
	return getArticleList(p.baseURL, board, n-pages, n)
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
	return left
}

func getArticleList(baseURL, board string, start, end int) (articleList []string, err error) {
	for i := start; i <= end; i++ {
		doc, err := parseUrl(baseURL + "bbs/" + board + "/index" + strconv.Itoa(i) + ".html")
		if err != nil {
			return nil, err
		}
		doc.Find(".title").Each(func(i int, s *goquery.Selection) {
			val, exist := s.Children().Attr("href")
			if exist {
				articleList = append(articleList, baseURL+val[1:])
			}
		})
	}
	return
}
func checkArticlePage(url string, n int) bool {
	url = url + strconv.Itoa(n) + ".html"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.AddCookie(over18cookie)
	resp, err := defaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
