package ptt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	. "strings"

	"github.com/PuerkitoBio/goquery"
	mapset "github.com/deckarep/golang-set"
)

// crawlBoard
//

type PTT struct {
	baseUrl   string
	storePath string
	board     mapset.Set
}

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

func NewPTT(storePathFolder string) *PTT {
	p := new(PTT)
	p.baseUrl = "https://www.ptt.cc/bbs/"
	p.storePath = storePathFolder
	p.board = mapset.NewSet()
	return p
}

func (p *PTT) SetBoard(board string) error {
	if isValidBoard(p.baseUrl, board) {
		p.board.Add(board)
	} else {
		return fmt.Errorf("board name %s not valid", board)
	}
	return nil
}

func (p *PTT) SetBoardWithSlice(board []string) error {
	errMsg := ""
	for id := range board {
		if isValidBoard(p.baseUrl, board[id]) {
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

func isValidBoard(baseurl, board string) bool {
	if resp, err := http.Get(baseurl + board + "/index.html"); err != nil {
		fmt.Println(err)
		return false
	} else if resp.StatusCode != 200 {
		return false
	}
	return true
}

func CrawlArticle(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "over18", Value: "1"})
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("CrawlArticle: status code %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	author, board, class, title, timestamp := parseArticleAttr(doc.Find(".article-meta-value").Remove())
	doc.Find(".article-metaline").Remove()
	doc.Find(".article-metaline-right").Remove()

	// remove the replies if they're in the content
	content := Split(doc.Find("#main-content").Text(), "\n※ 發信站")[0]
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
		replies = append(replies, reply{
			Floor:     i + 1,
			UserID:    s.Find(".push-userid").Text(),
			Vote:      s.Find(".push-tag").Text(),
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
	f, err := os.Create("out.json")
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(jsondata)
	return nil
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
