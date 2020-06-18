package ptt

import (
	"encoding/json"
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

func CrawlArticle(url string) ([]byte, error) {
	doc, err := parseArticle(url)
	if err != nil {
		return nil, err
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
		replies = append(replies, reply{
			Floor:     i + 1,
			UserID:    TrimSpace(userid),
			Vote:      vote[:len(vote)-1],                 // remove space
			Content:   s.Find(".push-content").Text()[2:], // remove : and space
			Timestamp: ipdatetime[1 : len(ipdatetime)-1],  // remove space and newline
		})
		s.Remove()
	})
	// now "doc.Find("#main-content").Text()" has all the content including the author's reply in the reply area
	art := article{Board: board, Class: class, Title: title, Author: author, Timestamp: timestamp, Content: content, Replies: replies}
	jsondata, err := json.Marshal(art)
	if err != nil {
		log.Fatal(err)
	}
	return jsondata, nil
}

func parseArticle(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "over18", Value: "1"})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("CrawlArticle: status code %d %s", resp.StatusCode, resp.Status)
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
	for i := range artTitle {
		if artTitle[i] == '[' {
			for artTitle[i] != ']' {
				class = append(class, artTitle[i])
				i++
			}
			class = append(class, artTitle[i])
			break
		}
	}
	return TrimSpace(string(class)), TrimSpace(Trim(artTitle, string(class)))
}
