package ptt

import (
	"fmt"
	"log"
	"net/http"
	. "strings"

	"github.com/PuerkitoBio/goquery"
)

type reply struct {
	Floor     int    `json:"floor"`
	UserID    string `json:"author"`
	Vote      string `json:"vote"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type article struct {
	Board     string  `json:"Board"`
	Class     string  `json:"Class"`
	Title     string  `json:"Title"`
	Author    string  `json:"Author"`
	Timestamp string  `json:"Timestamp"`
	Content   string  `json:"Content"`
	Replies   []reply `json:"Replies"`
}

func CrawlArticle(url string) (article, error) {
	return getArticle(url)
}

func CrawlArticleThread(url string, ch chan article, sem chan int) {
	a, err := getArticle(url)
	if err == nil {
		ch <- a
	}
	<-sem

}
func getArticle(url string) (article, error) {
	doc, err := parseUrl(url)
	if err != nil {
		log.Println(err)
		return article{}, err
	}
	author, board, class, title, timestamp := parseArticleAttr(doc.Find(".article-meta-value").Remove())
	doc.Find(".article-metaline").Remove()
	doc.Find(".article-metaline-right").Remove()

	content := Split(doc.Find("#main-content").Text(), "\n※ 發信站")[0]

	// remove the replies if they're in the content which means they're part of the signature
	doc.Find(".push").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if Contains(content, s.Text()) {
			s.Remove()
		}
		if i > 4 {
			return false
		}
		return true
	})
	replies := make([]reply, 0)
	doc.Find(".push").Each(func(i int, s *goquery.Selection) {
		ipdatetime := s.Find(".push-ipdatetime").Text()
		vote := s.Find(".push-tag").Text()
		userid := s.Find(".push-userid").Text()
		trimContent := func(s string) string {
			if s[0] == ':' {
				return TrimSpace(s[1:])
			}
			return TrimSpace(s)
		}
		replies = append(replies, reply{
			Floor:     i + 1,
			UserID:    TrimSpace(userid),
			Vote:      TrimSpace(vote), // remove space
			Content:   trimContent(s.Find(".push-content").Text()),
			Timestamp: TrimSpace(ipdatetime), // remove space and newline
		})
		s.Remove()
	})
	// now "doc.Find("#main-content").Text()" has all the content including the author's reply in the reply area
	return article{Board: board, Class: class, Title: title, Author: author, Timestamp: timestamp, Content: content, Replies: replies}, nil
}
func parseUrl(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(over18cookie)
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ParseUrl: status code %d %s at %s url", resp.StatusCode, resp.Status, url)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
func parseArticleAttr(s *goquery.Selection) (author, board, class, title, date string) {
	author = TrimSpace(Split(s.Eq(0).Text(), "(")[0])
	board = s.Eq(1).Text()
	class, title = parseTitle(s.Eq(2).Text())
	date = s.Eq(3).Text()
	return author, board, class, title, date
}
func parseTitle(artTitle string) (string, string) {
	var class []byte
	classStr := ""
	for i := range artTitle {
		if artTitle[i] == '[' {
			for i < len(artTitle) && artTitle[i] != ']' {
				class = append(class, artTitle[i])
				i++
			}
			if i != len(artTitle) {
				class = append(class, artTitle[i])
				classStr = string(class)
			} else {
				classStr = ""
			}
			break
		}
	}
	return TrimSpace(classStr), TrimSpace(Trim(artTitle, classStr))
}
